'use client'

import { useState, useEffect } from 'react'
import { getWebSocketClient, type WebSocketMessage } from '@/lib/websocket'

interface Agent {
  key: string
  name: string
  status: 'idle' | 'thinking' | 'responding' | 'error'
  currentTask?: string
  lastUpdate?: string
}

export default function AgentStatus() {
  const [agents, setAgents] = useState<Agent[]>([
    { key: 'CD', name: 'Creative Director', status: 'idle' },
    { key: 'SA', name: 'Strategy Analyst', status: 'idle' },
    { key: 'TD', name: 'Technical Director', status: 'idle' },
    { key: 'CE', name: 'Chief Editor', status: 'idle' },
    { key: 'AD', name: 'Art Director', status: 'idle' }
  ])

  useEffect(() => {
    const ws = getWebSocketClient()

    const unsubscribe = ws.onMessage((message: WebSocketMessage) => {
      if (message.type === 'agent_selected') {
        // 선택된 에이전트 상태 업데이트
        const agentKey = message.agent
        setAgents(prev => prev.map(agent =>
          agent.key === agentKey
            ? { ...agent, status: 'thinking' as const }
            : { ...agent, status: 'idle' as const }
        ))
      } else if (message.type === 'agent_response') {
        // 응답 완료
        const agentKey = message.agent
        setAgents(prev => prev.map(agent =>
          agent.key === agentKey
            ? { ...agent, status: 'idle' as const, lastUpdate: new Date().toISOString() }
            : agent
        ))
      } else if (message.type === 'agent_error') {
        // 에러 발생
        setAgents(prev => prev.map(agent => ({
          ...agent,
          status: 'idle' as const
        })))
      }
    })

    return () => {
      unsubscribe()
    }
  }, [])

  const getStatusColor = (status: Agent['status']) => {
    switch (status) {
      case 'thinking':
        return 'bg-yellow-500 animate-pulse'
      case 'responding':
        return 'bg-green-500 animate-pulse'
      case 'error':
        return 'bg-red-500'
      default:
        return 'bg-brand-gray-300'
    }
  }

  const getAgentIcon = (key: string) => {
    const icons: Record<string, string> = {
      CD: '👑',
      SA: '📊',
      TD: '⚙️',
      CE: '✍️',
      AD: '🎨'
    }
    return icons[key] || '🤖'
  }

  const getAgentColor = (key: string) => {
    const colors: Record<string, string> = {
      CD: 'border-purple-200 bg-purple-50',
      SA: 'border-blue-200 bg-blue-50',
      TD: 'border-green-200 bg-green-50',
      CE: 'border-orange-200 bg-orange-50',
      AD: 'border-pink-200 bg-pink-50'
    }
    return colors[key] || 'border-brand-gray-200 bg-brand-gray-50'
  }

  return (
    <div className="bg-white rounded-2xl p-6 subtle-shadow">
      <h3 className="text-lg font-semibold text-brand-black mb-4">
        에이전트 상태
      </h3>

      <div className="space-y-3">
        {agents.map((agent) => (
          <div
            key={agent.key}
            className={`p-4 rounded-xl border-2 transition-all ${getAgentColor(agent.key)} ${
              agent.status === 'thinking' ? 'ring-2 ring-brand-gold' : ''
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="text-2xl">{getAgentIcon(agent.key)}</div>
                <div>
                  <div className="font-semibold text-brand-black text-sm">
                    {agent.name}
                  </div>
                  <div className="text-xs text-brand-gray-500 font-mono">
                    {agent.key}
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className={`w-3 h-3 rounded-full ${getStatusColor(agent.status)}`} />
                <span className="text-xs text-brand-gray-600 capitalize">
                  {agent.status === 'thinking' ? 'Working' :
                   agent.status === 'responding' ? 'Responding' :
                   agent.status === 'error' ? 'Error' : 'Idle'}
                </span>
              </div>
            </div>

            {agent.status === 'thinking' && (
              <div className="mt-3 pt-3 border-t border-current border-opacity-20">
                <div className="flex items-center gap-2">
                  <div className="flex gap-1">
                    <div className="w-1.5 h-1.5 bg-brand-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                    <div className="w-1.5 h-1.5 bg-brand-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                    <div className="w-1.5 h-1.5 bg-brand-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                  </div>
                  <span className="text-xs text-brand-gray-600">Processing request...</span>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="mt-4 pt-4 border-t border-brand-gray-100">
        <p className="text-xs text-brand-gray-500">
          에이전트는 메시지 키워드를 분석하여 자동으로 선택됩니다
        </p>
      </div>
    </div>
  )
}
