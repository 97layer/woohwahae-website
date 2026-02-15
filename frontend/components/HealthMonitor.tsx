'use client'

import { useEffect, useState } from 'react'
import { getWebSocketClient, type SyncState, type WebSocketMessage } from '@/lib/websocket'

export default function HealthMonitor() {
  const [syncState, setSyncState] = useState<SyncState | null>(null)
  const [wsStatus, setWsStatus] = useState<'disconnected' | 'connecting' | 'connected'>('disconnected')
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)

  useEffect(() => {
    const ws = getWebSocketClient()

    // WebSocket 연결
    ws.connect()
      .then(() => {
        setWsStatus('connected')
        // 연결 후 상태 요청
        ws.send({ type: 'get_status' })
      })
      .catch((error) => {
        console.error('Failed to connect:', error)
        setWsStatus('disconnected')
      })

    // 메시지 핸들러 등록
    const unsubscribe = ws.onMessage((message: WebSocketMessage) => {
      if (message.type === 'sync_state_update' && message.data) {
        setSyncState(message.data)
        setLastUpdate(new Date())
      }
    })

    // 주기적으로 연결 상태 체크
    const statusInterval = setInterval(() => {
      const status = ws.getStatus()
      if (status === 'open') {
        setWsStatus('connected')
      } else if (status === 'connecting') {
        setWsStatus('connecting')
      } else {
        setWsStatus('disconnected')
      }
    }, 1000)

    return () => {
      unsubscribe()
      clearInterval(statusInterval)
      // ws.disconnect() // 페이지 이동 시에도 유지
    }
  }, [])

  const getNodeStatusColor = (health: string) => {
    switch (health) {
      case 'online':
        return 'bg-green-500'
      case 'offline':
        return 'bg-red-500'
      default:
        return 'bg-gray-400'
    }
  }

  const formatTimestamp = (isoString: string | null) => {
    if (!isoString) return 'Never'
    const date = new Date(isoString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffSec = Math.floor(diffMs / 1000)

    if (diffSec < 60) return `${diffSec}초 전`
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)}분 전`
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}시간 전`
    return date.toLocaleString('ko-KR')
  }

  return (
    <div className="min-h-screen bg-brand-white p-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <header className="mb-12">
          <h1 className="text-4xl font-bold text-brand-black mb-2">
            97layerOS
          </h1>
          <p className="text-brand-gray-500 text-sm">
            Strategic Intelligence Display
          </p>
        </header>

        {/* WebSocket Status Badge */}
        <div className="mb-8 flex items-center gap-2">
          <div className={`w-3 h-3 rounded-full ${
            wsStatus === 'connected' ? 'bg-green-500 animate-pulse' :
            wsStatus === 'connecting' ? 'bg-yellow-500 animate-pulse' :
            'bg-red-500'
          }`} />
          <span className="text-sm text-brand-gray-600">
            {wsStatus === 'connected' ? 'Real-time Connected' :
             wsStatus === 'connecting' ? 'Connecting...' :
             'Disconnected'}
          </span>
        </div>

        {/* Hybrid Health Monitor */}
        <div className="bg-white rounded-2xl p-8 subtle-shadow mb-6">
          <h2 className="text-2xl font-semibold text-brand-black mb-6">
            하이브리드 상태 모니터
          </h2>

          {!syncState ? (
            <div className="text-center py-12 text-brand-gray-400">
              <div className="animate-spin w-8 h-8 border-4 border-brand-gray-200 border-t-brand-gold rounded-full mx-auto mb-4" />
              <p>Loading system state...</p>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Active Node Indicator */}
              <div className="flex items-center justify-between p-6 bg-brand-gray-50 rounded-xl">
                <div>
                  <div className="text-sm text-brand-gray-500 mb-1">Active Node</div>
                  <div className="text-2xl font-bold text-brand-black">
                    {syncState.active_node === 'macbook' ? '🖥️ MacBook' : '☁️ GCP VM'}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-sm text-brand-gray-500 mb-1">Location</div>
                  <div className="text-lg font-medium text-brand-gray-700">
                    {syncState.location}
                  </div>
                </div>
              </div>

              {/* Node Health Status */}
              <div className="grid grid-cols-2 gap-4">
                {/* MacBook */}
                <div className="p-6 bg-white border-2 border-brand-gray-100 rounded-xl">
                  <div className="flex items-center gap-3 mb-3">
                    <div className={`w-4 h-4 rounded-full ${getNodeStatusColor(syncState.health.macbook)}`} />
                    <span className="font-semibold text-brand-black">MacBook</span>
                  </div>
                  <div className="text-sm text-brand-gray-600">
                    Status: <span className="font-medium">{syncState.health.macbook}</span>
                  </div>
                  {syncState.active_node === 'macbook' && (
                    <div className="mt-2 text-xs text-brand-gold font-medium">
                      ● ACTIVE
                    </div>
                  )}
                </div>

                {/* GCP VM */}
                <div className="p-6 bg-white border-2 border-brand-gray-100 rounded-xl">
                  <div className="flex items-center gap-3 mb-3">
                    <div className={`w-4 h-4 rounded-full ${getNodeStatusColor(syncState.health.gcp_vm)}`} />
                    <span className="font-semibold text-brand-black">GCP VM</span>
                  </div>
                  <div className="text-sm text-brand-gray-600">
                    Status: <span className="font-medium">{syncState.health.gcp_vm}</span>
                  </div>
                  {syncState.active_node === 'gcp_vm' && (
                    <div className="mt-2 text-xs text-brand-gold font-medium">
                      ● ACTIVE
                    </div>
                  )}
                </div>
              </div>

              {/* Sync Information */}
              <div className="pt-4 border-t border-brand-gray-100">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-brand-gray-500">Last Sync:</span>
                    <p className="font-medium text-brand-gray-800 mt-1">
                      {formatTimestamp(syncState.last_sync)}
                    </p>
                  </div>
                  <div>
                    <span className="text-brand-gray-500">Last Heartbeat:</span>
                    <p className="font-medium text-brand-gray-800 mt-1">
                      {formatTimestamp(syncState.last_heartbeat)}
                    </p>
                  </div>
                </div>

                {syncState.pending_handover && (
                  <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                    <span className="text-sm text-yellow-800">
                      ⚠️ Handover pending
                    </span>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Real-time Update Indicator */}
        {lastUpdate && (
          <div className="text-center text-sm text-brand-gray-400">
            Last updated: {lastUpdate.toLocaleTimeString('ko-KR')}
          </div>
        )}

        {/* Phase 1 MVP Notice */}
        <div className="mt-12 p-6 bg-brand-black text-brand-white rounded-xl">
          <h3 className="font-semibold mb-2">Phase 1: 메신저 뼈대 완성</h3>
          <p className="text-sm text-brand-gray-300">
            ✅ FastAPI + WebSocket 실시간 통신<br />
            ✅ Hybrid Health Monitor (MacBook ↔ GCP)<br />
            🔜 Phase 2: 에이전트 오케스트레이션 채팅
          </p>
        </div>
      </div>
    </div>
  )
}
