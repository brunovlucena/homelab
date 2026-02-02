import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Gas Station Command Center | POS Edge',
  description: '⛽ Real-time monitoring and control for gas station operations',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-fuel-black bg-fuel-grid bg-grid">
        {children}
      </body>
    </html>
  )
}
