import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: '97layerOS - Strategic Intelligence Display',
  description: 'Real-time hybrid AI orchestration system',
  manifest: '/manifest.json',
  themeColor: '#0A0A0A',
  appleWebApp: {
    capable: true,
    statusBarStyle: 'black-translucent',
    title: '97layerOS'
  }
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ko">
      <body className={inter.className}>{children}</body>
    </html>
  )
}
