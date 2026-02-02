'use client'

import { Bell, AlertTriangle, Info, CheckCircle } from 'lucide-react'

export function AlertsView() {
  return (
    <div className="space-y-6">
      <div className="card p-6">
        <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
          <Bell className="w-5 h-5 text-medical-blue" />
          System Alerts
        </h3>

        <div className="space-y-3">
          <div className="flex items-start gap-3 p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
            <AlertTriangle className="w-5 h-5 text-yellow-500 mt-0.5" />
            <div className="flex-1">
              <p className="font-medium text-yellow-500">Backend Connection Issue</p>
              <p className="text-sm text-gray-400 mt-1">
                Unable to connect to agent-medical backend. Check if the service is running.
              </p>
              <p className="text-xs text-gray-500 mt-2">Just now</p>
            </div>
          </div>

          <div className="flex items-start gap-3 p-4 bg-medical-blue/10 border border-medical-blue/30 rounded-lg">
            <Info className="w-5 h-5 text-medical-blue mt-0.5" />
            <div className="flex-1">
              <p className="font-medium text-medical-blue">HIPAA Compliance Check</p>
              <p className="text-sm text-gray-400 mt-1">
                Monthly compliance audit scheduled for tomorrow
              </p>
              <p className="text-xs text-gray-500 mt-2">2 hours ago</p>
            </div>
          </div>

          <div className="flex items-start gap-3 p-4 bg-medical-green/10 border border-medical-green/30 rounded-lg">
            <CheckCircle className="w-5 h-5 text-medical-green mt-0.5" />
            <div className="flex-1">
              <p className="font-medium text-medical-green">System Healthy</p>
              <p className="text-sm text-gray-400 mt-1">
                All systems operational. HIPAA compliance: 98%
              </p>
              <p className="text-xs text-gray-500 mt-2">5 hours ago</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
