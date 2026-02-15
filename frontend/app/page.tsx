'use client'

import { useState } from 'react'
import HealthMonitor from '@/components/HealthMonitor'
import AgentChat from '@/components/AgentChat'
import AgentStatus from '@/components/AgentStatus'

export default function Home() {
  const [activeView, setActiveView] = useState<'health' | 'chat'>('health')
  const userId = 'pwa_user' // TODO: Phase 4에서 OAuth2 인증 후 실제 user ID 사용

  return (
    <div className="min-h-screen bg-brand-white">
      {/* Header with Navigation */}
      <div className="bg-white border-b border-brand-gray-100">
        <div className="max-w-7xl mx-auto px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-brand-black">97layerOS</h1>
              <p className="text-sm text-brand-gray-500">Strategic Intelligence Display</p>
            </div>

            {/* View Toggle */}
            <div className="flex gap-2 bg-brand-gray-50 p-1 rounded-xl">
              <button
                onClick={() => setActiveView('health')}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                  activeView === 'health'
                    ? 'bg-brand-black text-brand-white'
                    : 'text-brand-gray-600 hover:text-brand-black'
                }`}
              >
                🏥 Health Monitor
              </button>
              <button
                onClick={() => setActiveView('chat')}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                  activeView === 'chat'
                    ? 'bg-brand-black text-brand-white'
                    : 'text-brand-gray-600 hover:text-brand-black'
                }`}
              >
                💬 Agent Chat
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-8 py-8">
        {activeView === 'health' ? (
          <HealthMonitor />
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Chat Area - 2/3 width on large screens */}
            <div className="lg:col-span-2 h-[calc(100vh-200px)]">
              <AgentChat userId={userId} />
            </div>

            {/* Agent Status Sidebar - 1/3 width on large screens */}
            <div className="lg:col-span-1">
              <AgentStatus />
            </div>
          </div>
        )}
      </div>

      {/* Phase Indicator */}
      <div className="fixed bottom-6 right-6">
        <div className="bg-brand-black text-brand-white px-4 py-2 rounded-full text-xs font-medium shadow-lg">
          Phase 2 🚧 Agent Chat
        </div>
      </div>
    </div>
  )
}
