export interface Patient {
  id: string
  name: string
  dateOfBirth: string
  email?: string
  phone?: string
  status: 'active' | 'inactive'
}

export interface MedicalRecord {
  id: string
  patientId: string
  type: 'lab' | 'prescription' | 'diagnosis' | 'visit'
  date: string
  title: string
  description: string
}

export interface Agent {
  name: string
  status: 'online' | 'offline'
  version: string
  metrics?: {
    cpu: number
    memory: number
    requests: number
  }
}

export interface Metrics {
  totalPatients: number
  activePatients: number
  totalRecords: number
  recordsLast24h: number
  queriesLast24h: number
  hipaaAudits: number
  agentStatus: 'online' | 'offline'
  complianceScore: number
}
