import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Doctor {
  doctor_id: string;
  name: string;
  email: string;
  crm: string;
  specialization?: string;
}

export interface Patient {
  patient_id: string;
  name: string;
  cpf?: string;
  birth_date?: string;
}

export interface CaseSummary {
  summary: string;
  patient_id: string;
  case_id?: string;
  created_at: string;
}

export interface PatientRecord {
  record_id: string;
  patient_id: string;
  type: 'exam' | 'lab_result' | 'prescription' | 'consultation';
  content: any;
  date: string;
}

export interface CorrelationResult {
  correlation: any;
  patient_id: string;
  query: string;
  insights: string[];
}

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const medicalService = {
  // Summarize a case
  summarizeCase: async (
    doctorId: string,
    patientId: string,
    caseId?: string
  ): Promise<CaseSummary> => {
    const response = await api.post('/summarize', {
      doctor_id: doctorId,
      patient_id: patientId,
      case_id: caseId,
    });
    return response.data.data;
  },

  // Correlate medical data
  correlateData: async (
    doctorId: string,
    patientId: string,
    query: string
  ): Promise<CorrelationResult> => {
    const response = await api.post('/correlate', {
      doctor_id: doctorId,
      patient_id: patientId,
      query,
    });
    return response.data.data;
  },

  // Get patient records
  getPatientRecords: async (
    doctorId: string,
    patientId: string,
    recordType?: string
  ): Promise<PatientRecord[]> => {
    const response = await api.post('/patient-records', {
      doctor_id: doctorId,
      patient_id: patientId,
      record_type: recordType,
    });
    return response.data.data.records || [];
  },

  // Send alert
  sendAlert: async (
    doctorId: string,
    alertType: string,
    message: string,
    patientId?: string
  ): Promise<void> => {
    await api.post('/alert', {
      doctor_id: doctorId,
      alert_type: alertType,
      message,
      patient_id: patientId,
    });
  },
};

export default api;
