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

export function formatLiters(value: number): string {
  return new Intl.NumberFormat('pt-BR', {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1
  }).format(value) + ' L'
}

export function formatPercent(value: number, total: number): number {
  return Math.round((value / total) * 100)
}

export function getTimeSince(date: string): string {
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000)
  
  if (seconds < 60) return 'agora'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m atrás`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h atrás`
  return `${Math.floor(seconds / 86400)}d atrás`
}

export function getFuelColor(fuelType: string): string {
  switch (fuelType) {
    case 'gasoline': return 'fuel-green'
    case 'diesel': return 'diesel'
    case 'premium': return 'premium'
    case 'ethanol': return 'fuel-lime'
    default: return 'fuel-gray'
  }
}

export function getTankStatusColor(status: string): string {
  switch (status) {
    case 'normal': return 'fuel-green'
    case 'low': return 'fuel-amber'
    case 'critical': return 'fuel-red'
    case 'refilling': return 'fuel-blue'
    default: return 'fuel-gray'
  }
}
