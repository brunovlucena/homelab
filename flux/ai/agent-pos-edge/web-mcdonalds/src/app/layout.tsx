import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: "McDonald's Command Center | Kitchen & POS",
  description: "🍔 Real-time kitchen display and order management system",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-mc-black bg-mc-pattern">
        {children}
      </body>
    </html>
  )
}
