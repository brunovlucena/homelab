'use client'

import { FileText, Search } from 'lucide-react'

export function RecordsView() {
  return (
    <div className="space-y-6">
      <div className="card p-6">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-bold flex items-center gap-2">
            <FileText className="w-5 h-5 text-medical-blue" />
            Medical Records
          </h3>
          <button className="btn-primary">Create Record</button>
        </div>
        
        <div className="mb-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Search medical records..."
              className="input pl-10"
            />
          </div>
        </div>

        <div className="text-center py-12 text-gray-500">
          <FileText className="w-16 h-16 mx-auto mb-4 opacity-50" />
          <p>No medical records found</p>
          <p className="text-sm mt-1">Connect to medical agent backend to see records</p>
        </div>
      </div>
    </div>
  )
}
