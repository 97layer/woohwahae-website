'use client'

import { useState, useEffect, useRef } from 'react'
import { getWebSocketClient, type WebSocketMessage } from '@/lib/websocket'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  agent?: string
  agent_name?: string
  timestamp: string
}

interface AgentChatProps {
  userId: string
}

export default function AgentChat({ userId }: AgentChatProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [inputValue, setInputValue] = useState('')
  const [isAgentThinking, setIsAgentThinking] = useState(false)
  const [currentAgent, setCurrentAgent] = useState<string | null>(null)
  const [thinkingMessage, setThinkingMessage] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const ws = getWebSocketClient()

    // WebSocket 메시지 핸들러
    const unsubscribe = ws.onMessage((message: WebSocketMessage) => {
      if (message.type === 'agent_thinking') {
        setIsAgentThinking(true)
        setThinkingMessage(message.data?.message || '에이전트가 사고 중...')
      } else if (message.type === 'agent_selected') {
        setCurrentAgent(message.agent || null)
        setThinkingMessage(`${message.data?.agent_name || message.agent} 응답 생성 중...`)
      } else if (message.type === 'agent_response') {
        setIsAgentThinking(false)
        setCurrentAgent(null)

        // 에이전트 응답 추가
        const newMessage: Message = {
          id: Date.now().toString(),
          role: 'assistant',
          content: message.data?.message || message.message || '',
          agent: message.agent,
          agent_name: message.data?.agent_name || message.agent_name,
          timestamp: message.data?.timestamp || message.timestamp || new Date().toISOString()
        }
        setMessages(prev => [...prev, newMessage])
      } else if (message.type === 'agent_error') {
        setIsAgentThinking(false)
        setCurrentAgent(null)

        const errorMessage: Message = {
          id: Date.now().toString(),
          role: 'system',
          content: `오류: ${message.data?.error || message.error || '알 수 없는 오류'}`,
          timestamp: new Date().toISOString()
        }
        setMessages(prev => [...prev, errorMessage])
      }
    })

    // 대화 기록 로드
    loadChatHistory()

    return () => {
      unsubscribe()
    }
  }, [userId])

  // 메시지 추가 시 자동 스크롤
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const loadChatHistory = async () => {
    try {
      const response = await fetch(`http://localhost:8080/api/chat/history/${userId}?limit=20`)
      if (response.ok) {
        const data = await response.json()
        const loadedMessages: Message[] = data.messages.map((msg: any, idx: number) => ({
          id: `history-${idx}`,
          role: msg.role,
          content: msg.content,
          timestamp: msg.timestamp
        }))
        setMessages(loadedMessages)
      }
    } catch (error) {
      console.error('Failed to load chat history:', error)
    }
  }

  const sendMessage = async () => {
    if (!inputValue.trim()) return

    // 사용자 메시지 추가
    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue,
      timestamp: new Date().toISOString()
    }
    setMessages(prev => [...prev, userMessage])
    setInputValue('')
    setIsAgentThinking(true)
    setThinkingMessage('에이전트 선택 중...')

    try {
      // WebSocket을 통해 전송 (실시간)
      const ws = getWebSocketClient()
      ws.send({
        type: 'chat',
        user_id: userId,
        message: inputValue
      })
    } catch (error) {
      console.error('Failed to send message:', error)
      setIsAgentThinking(false)
    }
  }

  const getAgentColor = (agent?: string) => {
    const colors: Record<string, string> = {
      CD: 'text-purple-600',
      SA: 'text-blue-600',
      TD: 'text-green-600',
      CE: 'text-orange-600',
      AD: 'text-pink-600'
    }
    return agent ? colors[agent] || 'text-brand-gray-600' : 'text-brand-gray-600'
  }

  const getAgentBadge = (agent?: string, agent_name?: string) => {
    if (!agent) return null
    return (
      <span className={`text-xs font-semibold px-2 py-1 rounded ${getAgentColor(agent)} bg-opacity-10 bg-current`}>
        {agent_name || agent}
      </span>
    )
  }

  const formatTimestamp = (isoString: string) => {
    const date = new Date(isoString)
    return date.toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className="flex flex-col h-full bg-white rounded-2xl subtle-shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-brand-gray-100">
        <h2 className="text-xl font-semibold text-brand-black">에이전트 오케스트레이션</h2>
        <p className="text-sm text-brand-gray-500 mt-1">
          에이전트와 대화하고 전략적 결정을 내리세요
        </p>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        {messages.length === 0 ? (
          <div className="text-center py-12 text-brand-gray-400">
            <p className="text-lg mb-2">💬</p>
            <p>대화를 시작하세요</p>
            <p className="text-sm mt-1">에이전트가 자동으로 선택됩니다</p>
          </div>
        ) : (
          messages.map((msg) => (
            <div
              key={msg.id}
              className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-[70%] ${
                  msg.role === 'user'
                    ? 'bg-brand-black text-brand-white'
                    : msg.role === 'system'
                    ? 'bg-red-50 text-red-800 border border-red-200'
                    : 'bg-brand-gray-50 text-brand-black'
                } rounded-2xl px-4 py-3`}
              >
                {msg.agent && (
                  <div className="mb-2">
                    {getAgentBadge(msg.agent, msg.agent_name)}
                  </div>
                )}
                <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                <p className={`text-xs mt-2 ${
                  msg.role === 'user' ? 'text-brand-gray-300' : 'text-brand-gray-500'
                }`}>
                  {formatTimestamp(msg.timestamp)}
                </p>
              </div>
            </div>
          ))
        )}

        {/* Agent Thinking Indicator */}
        {isAgentThinking && (
          <div className="flex justify-start">
            <div className="max-w-[70%] bg-brand-gray-50 rounded-2xl px-4 py-3">
              <div className="flex items-center gap-2">
                <div className="flex gap-1">
                  <div className="w-2 h-2 bg-brand-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                  <div className="w-2 h-2 bg-brand-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                  <div className="w-2 h-2 bg-brand-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                </div>
                <span className="text-sm text-brand-gray-600">{thinkingMessage}</span>
              </div>
              {currentAgent && (
                <div className="mt-2">
                  {getAgentBadge(currentAgent)}
                </div>
              )}
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="px-6 py-4 border-t border-brand-gray-100">
        <div className="flex gap-2">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
            placeholder="메시지를 입력하세요..."
            disabled={isAgentThinking}
            className="flex-1 px-4 py-3 border border-brand-gray-200 rounded-xl focus:outline-none focus:border-brand-gold disabled:bg-brand-gray-50 disabled:cursor-not-allowed"
          />
          <button
            onClick={sendMessage}
            disabled={!inputValue.trim() || isAgentThinking}
            className="px-6 py-3 bg-brand-black text-brand-white rounded-xl font-medium hover:bg-brand-gray-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            전송
          </button>
        </div>
        <p className="text-xs text-brand-gray-400 mt-2">
          키워드에 따라 자동으로 적절한 에이전트가 선택됩니다
        </p>
      </div>
    </div>
  )
}
