'use client'

import { Users, Search } from 'lucide-react'

export function PatientsView() {
  return (
    <div className="space-y-6">
      <div className="card p-6">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-bold flex items-center gap-2">
            <Users className="w-5 h-5 text-medical-blue" />
            Patient Management
          </h3>
          <button className="btn-primary">Add New Patient</button>
        </div>
        
        <div className="mb-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Search patients by name, ID, or record..."
              className="input pl-10"
            />
          </div>
        </div>

        <div className="text-center py-12 text-gray-500">
          <Users className="w-16 h-16 mx-auto mb-4 opacity-50" />
          <p>No patients found</p>
          <p className="text-sm mt-1">Connect to medical agent backend to see patient data</p>
        </div>
      </div>
    </div>
  )
}
