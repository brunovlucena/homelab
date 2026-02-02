'use client'

import { Shield, CheckCircle, AlertCircle, FileText, Lock } from 'lucide-react'

export function ComplianceView() {
  const complianceItems = [
    { label: 'Data Encryption', status: 'compliant', description: 'All data encrypted at rest and in transit' },
    { label: 'Access Control', status: 'compliant', description: 'Role-based access control (RBAC) enabled' },
    { label: 'Audit Logging', status: 'compliant', description: 'All access events logged and monitored' },
    { label: 'Data Retention', status: 'compliant', description: '7-year retention policy active' },
    { label: 'Patient Privacy', status: 'compliant', description: 'PHI isolation and masking enforced' },
    { label: 'Backup & Recovery', status: 'warning', description: 'Last backup: 2 hours ago' },
  ]

  return (
    <div className="space-y-6">
      {/* HIPAA Compliance Overview */}
      <div className="card p-6">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-bold flex items-center gap-2">
            <Shield className="w-5 h-5 text-medical-blue" />
            HIPAA Compliance Dashboard
          </h3>
          <span className="badge-hipaa">Certified</span>
        </div>

        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="p-4 bg-medical-green/10 border border-medical-green/30 rounded-lg">
            <p className="text-sm text-gray-400">Compliance Score</p>
            <p className="text-3xl font-bold text-medical-green mt-1">98%</p>
          </div>
          <div className="p-4 bg-medical-blue/10 border border-medical-blue/30 rounded-lg">
            <p className="text-sm text-gray-400">Audits (30d)</p>
            <p className="text-3xl font-bold text-medical-blue mt-1">1,234</p>
          </div>
          <div className="p-4 bg-cyan-500/10 border border-cyan-500/30 rounded-lg">
            <p className="text-sm text-gray-400">Last Assessment</p>
            <p className="text-lg font-bold text-cyan-500 mt-1">Today</p>
          </div>
        </div>

        <div className="space-y-3">
          {complianceItems.map((item, index) => (
            <div key={index} className="flex items-center justify-between p-4 bg-cyber-dark/30 rounded-lg">
              <div className="flex items-center gap-3">
                {item.status === 'compliant' ? (
                  <CheckCircle className="w-5 h-5 text-medical-green" />
                ) : (
                  <AlertCircle className="w-5 h-5 text-yellow-500" />
                )}
                <div>
                  <p className="font-medium">{item.label}</p>
                  <p className="text-sm text-gray-500">{item.description}</p>
                </div>
              </div>
              <span className={item.status === 'compliant' ? 'badge-success' : 'badge-warning'}>
                {item.status === 'compliant' ? 'Compliant' : 'Review'}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Security Features */}
      <div className="grid grid-cols-2 gap-6">
        <div className="card p-6">
          <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
            <Lock className="w-5 h-5 text-medical-blue" />
            Security Features
          </h3>
          <ul className="space-y-2 text-sm">
            <li className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-medical-green" />
              <span>AES-256 encryption</span>
            </li>
            <li className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-medical-green" />
              <span>TLS 1.3 for data in transit</span>
            </li>
            <li className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-medical-green" />
              <span>JWT-based authentication</span>
            </li>
            <li className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-medical-green" />
              <span>Role-based access control (RBAC)</span>
            </li>
            <li className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-medical-green" />
              <span>Comprehensive audit logging</span>
            </li>
          </ul>
        </div>

        <div className="card p-6">
          <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
            <FileText className="w-5 h-5 text-medical-blue" />
            Audit Logs
          </h3>
          <div className="text-center py-8 text-gray-500">
            <FileText className="w-12 h-12 mx-auto mb-3 opacity-50" />
            <p className="text-sm">Connect to backend to view audit logs</p>
          </div>
        </div>
      </div>
    </div>
  )
}
