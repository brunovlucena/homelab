import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

export function formatTime(date: string | Date): string {
  return new Date(date).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function getTimeSince(date: string | Date): string {
  const now = new Date()
  const then = new Date(date)
  const diff = Math.floor((now.getTime() - then.getTime()) / 60000)
  
  if (diff < 1) return 'Just now'
  if (diff < 60) return `${diff}m ago`
  return `${Math.floor(diff / 60)}h ${diff % 60}m ago`
}
