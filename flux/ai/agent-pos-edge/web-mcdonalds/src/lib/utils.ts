import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  }).format(value)
}

export function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

export function getTimeSince(date: string): string {
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000)
  
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

export function getElapsedSeconds(date: string): number {
  return Math.floor((Date.now() - new Date(date).getTime()) / 1000)
}

export function getOrderTypeLabel(type: string): string {
  switch (type) {
    case 'dine-in': return 'Salão'
    case 'drive-thru': return 'Drive-Thru'
    case 'delivery': return 'Delivery'
    case 'takeaway': return 'Viagem'
    default: return type
  }
}

export function getOrderTypeColor(type: string): string {
  switch (type) {
    case 'dine-in': return 'mc-blue'
    case 'drive-thru': return 'mc-gold'
    case 'delivery': return 'mc-orange'
    case 'takeaway': return 'mc-green'
    default: return 'mc-gray'
  }
}

export function getStatusColor(status: string): string {
  switch (status) {
    case 'new': return 'status-new'
    case 'preparing': return 'status-preparing'
    case 'ready': return 'status-ready'
    case 'delivered': return 'status-delivered'
    default: return 'mc-gray'
  }
}
