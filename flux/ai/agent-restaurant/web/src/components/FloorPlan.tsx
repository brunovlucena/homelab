'use client'

import { useRestaurantStore, Table } from '@/store/restaurantStore'
import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Users, Clock, Wine } from 'lucide-react'
import { useState } from 'react'

export function FloorPlan() {
  const { tables, updateTable } = useRestaurantStore()
  const [selectedTable, setSelectedTable] = useState<Table | null>(null)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="font-serif text-2xl font-bold text-wine-900">Floor Plan</h1>
        <div className="flex gap-4">
          <Legend />
        </div>
      </div>
      
      <div className="grid grid-cols-3 gap-6">
        {/* Floor Plan Visualization */}
        <div className="col-span-2 elegant-card p-8 min-h-[600px] relative bg-cream-50">
          {/* Restaurant Layout */}
          <div className="absolute top-4 left-4 text-xs text-wood-400">ENTRANCE →</div>
          <div className="absolute top-4 right-4 text-xs text-wood-400">← KITCHEN</div>
          
          {/* Window Section */}
          <div className="absolute left-8 top-20 bottom-20 w-1 bg-cream-300 rounded" />
          <div className="absolute left-4 top-1/2 -translate-y-1/2 text-xs text-wood-400 -rotate-90">
            WINDOW
          </div>
          
          {/* Tables Grid */}
          <div className="grid grid-cols-3 gap-8 pt-12">
            {tables.map((table) => (
              <motion.div
                key={table.id}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setSelectedTable(table)}
                className={cn(
                  "relative aspect-[4/3] rounded-2xl cursor-pointer transition-all p-4 flex flex-col items-center justify-center",
                  table.status === 'available' && "bg-emerald-100 border-2 border-emerald-400 hover:border-emerald-500",
                  table.status === 'occupied' && "bg-wine-100 border-2 border-wine-400 hover:border-wine-500",
                  table.status === 'reserved' && "bg-gold-100 border-2 border-gold-400 hover:border-gold-500",
                  table.status === 'cleaning' && "bg-gray-100 border-2 border-gray-400",
                  selectedTable?.id === table.id && "ring-4 ring-wine-500/50"
                )}
              >
                {/* Table Number */}
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-1 bg-white rounded-full shadow-sm border border-cream-200">
                  <span className="font-mono text-sm font-bold">{table.id.replace('table-', '')}</span>
                </div>
                
                {/* Capacity Icons */}
                <div className="flex flex-wrap justify-center gap-1 mb-2">
                  {Array.from({ length: table.capacity }).map((_, i) => (
                    <div
                      key={i}
                      className={cn(
                        "w-6 h-6 rounded-full flex items-center justify-center text-xs",
                        table.status === 'occupied' && i < (table.partySize || 0)
                          ? "bg-wine-500 text-white"
                          : "bg-white border border-current opacity-50"
                      )}
                    >
                      👤
                    </div>
                  ))}
                </div>
                
                {/* Guest Info */}
                {table.guestName && (
                  <div className="text-center">
                    <p className="text-sm font-medium truncate max-w-full">{table.guestName}</p>
                    {table.server && (
                      <p className="text-xs text-wood-500">Server: {table.server}</p>
                    )}
                  </div>
                )}
                
                {table.status === 'available' && (
                  <span className="text-sm text-emerald-700 font-medium">Available</span>
                )}
                
                {table.status === 'reserved' && (
                  <span className="text-sm text-gold-700 font-medium">Reserved</span>
                )}
              </motion.div>
            ))}
          </div>
        </div>
        
        {/* Table Details Panel */}
        <div className="space-y-4">
          {selectedTable ? (
            <TableDetails 
              table={selectedTable} 
              onUpdate={(updates) => {
                updateTable(selectedTable.id, updates)
                setSelectedTable({ ...selectedTable, ...updates })
              }}
            />
          ) : (
            <div className="elegant-card p-6 text-center text-wood-400">
              <Users className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>Select a table to view details</p>
            </div>
          )}
          
          {/* Quick Stats */}
          <div className="elegant-card p-6">
            <h3 className="font-serif text-lg font-semibold text-wine-900 mb-4">Tonight's Stats</h3>
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-wood-500">Available</span>
                <span className="font-semibold text-emerald-600">
                  {tables.filter(t => t.status === 'available').length}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-wood-500">Occupied</span>
                <span className="font-semibold text-wine-600">
                  {tables.filter(t => t.status === 'occupied').length}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-wood-500">Reserved</span>
                <span className="font-semibold text-gold-600">
                  {tables.filter(t => t.status === 'reserved').length}
                </span>
              </div>
              <div className="border-t border-cream-200 pt-3 mt-3">
                <div className="flex justify-between">
                  <span className="text-wood-500">Total Capacity</span>
                  <span className="font-semibold">
                    {tables.reduce((sum, t) => sum + t.capacity, 0)} seats
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function TableDetails({ 
  table, 
  onUpdate 
}: { 
  table: Table
  onUpdate: (updates: Partial<Table>) => void 
}) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      className="elegant-card p-6"
    >
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-serif text-xl font-semibold text-wine-900">
          Table {table.id.replace('table-', '')}
        </h3>
        <span className={cn(
          "px-3 py-1 rounded-full text-sm font-medium capitalize",
          table.status === 'available' && "bg-emerald-100 text-emerald-700",
          table.status === 'occupied' && "bg-wine-100 text-wine-700",
          table.status === 'reserved' && "bg-gold-100 text-gold-700",
        )}>
          {table.status}
        </span>
      </div>
      
      <div className="space-y-4">
        <div className="flex items-center gap-3 text-wood-600">
          <Users className="w-5 h-5" />
          <span>Capacity: {table.capacity} guests</span>
        </div>
        
        {table.guestName && (
          <div className="p-4 bg-cream-50 rounded-lg">
            <p className="font-medium text-wine-900">{table.guestName}</p>
            {table.partySize && (
              <p className="text-sm text-wood-500">Party of {table.partySize}</p>
            )}
            {table.server && (
              <p className="text-sm text-wood-500">Server: {table.server}</p>
            )}
          </div>
        )}
        
        {/* Actions */}
        <div className="grid grid-cols-2 gap-2 pt-4 border-t border-cream-200">
          {table.status === 'available' && (
            <>
              <button 
                onClick={() => onUpdate({ status: 'occupied', guestName: 'Walk-in Guest' })}
                className="btn-elegant text-sm py-2"
              >
                Seat Guest
              </button>
              <button 
                onClick={() => onUpdate({ status: 'reserved' })}
                className="btn-elegant-outline text-sm py-2"
              >
                Reserve
              </button>
            </>
          )}
          {table.status === 'occupied' && (
            <>
              <button className="btn-elegant text-sm py-2">
                View Order
              </button>
              <button 
                onClick={() => onUpdate({ status: 'available', guestName: undefined, partySize: undefined })}
                className="btn-elegant-outline text-sm py-2"
              >
                Clear Table
              </button>
            </>
          )}
          {table.status === 'reserved' && (
            <button 
              onClick={() => onUpdate({ status: 'available', guestName: undefined })}
              className="btn-elegant-outline text-sm py-2 col-span-2"
            >
              Cancel Reservation
            </button>
          )}
        </div>
      </div>
    </motion.div>
  )
}

function Legend() {
  return (
    <div className="flex gap-4 text-sm">
      <span className="flex items-center gap-2">
        <div className="w-4 h-4 rounded bg-emerald-400" />
        Available
      </span>
      <span className="flex items-center gap-2">
        <div className="w-4 h-4 rounded bg-wine-400" />
        Occupied
      </span>
      <span className="flex items-center gap-2">
        <div className="w-4 h-4 rounded bg-gold-400" />
        Reserved
      </span>
    </div>
  )
}
