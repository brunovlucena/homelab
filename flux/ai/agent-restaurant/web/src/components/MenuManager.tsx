'use client'

import { useState } from 'react'
import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Plus, Edit, Wine, Leaf, AlertCircle } from 'lucide-react'

// Mock menu data
const menuCategories = [
  {
    name: 'Antipasti',
    items: [
      {
        id: 'bruschetta-trio',
        name: 'Bruschetta Trio',
        price: 18,
        description: 'Three artisan bruschettas with tomato basil, mushroom truffle, and burrata',
        dietary: ['vegetarian'],
        allergens: ['gluten', 'dairy'],
      },
      {
        id: 'carpaccio-manzo',
        name: 'Carpaccio di Manzo',
        price: 24,
        description: 'Thinly sliced beef tenderloin with arugula, capers, and aged parmesan',
        dietary: [],
        allergens: ['dairy'],
      },
    ],
  },
  {
    name: 'Primi',
    items: [
      {
        id: 'risotto-porcini',
        name: 'Risotto ai Porcini',
        price: 32,
        description: 'Carnaroli rice with wild porcini mushrooms and truffle oil',
        dietary: ['vegetarian', 'gluten-free'],
        allergens: ['dairy'],
        winePairing: '2020 Gavi di Gavi',
      },
      {
        id: 'tagliatelle-ragu',
        name: 'Tagliatelle al Ragù',
        price: 28,
        description: 'Fresh egg pasta with slow-cooked Bolognese sauce',
        dietary: [],
        allergens: ['gluten', 'dairy', 'eggs'],
        winePairing: '2018 Chianti Classico',
      },
    ],
  },
  {
    name: 'Secondi',
    items: [
      {
        id: 'branzino',
        name: 'Branzino alla Griglia',
        price: 42,
        description: 'Grilled Mediterranean sea bass with lemon and herbs',
        dietary: ['gluten-free'],
        allergens: ['fish'],
        winePairing: '2021 Vermentino',
      },
      {
        id: 'ossobuco',
        name: 'Ossobuco alla Milanese',
        price: 48,
        description: 'Braised veal shank with saffron risotto',
        dietary: ['gluten-free'],
        allergens: ['dairy'],
        winePairing: '2017 Barbaresco',
      },
    ],
  },
  {
    name: 'Dolci',
    items: [
      {
        id: 'tiramisu',
        name: 'Tiramisù della Casa',
        price: 14,
        description: 'Classic tiramisu with mascarpone and espresso',
        dietary: ['vegetarian'],
        allergens: ['dairy', 'eggs', 'gluten'],
      },
    ],
  },
]

export function MenuManager() {
  const [activeCategory, setActiveCategory] = useState(menuCategories[0].name)
  const [selectedItem, setSelectedItem] = useState<typeof menuCategories[0]['items'][0] | null>(null)

  const currentCategory = menuCategories.find(c => c.name === activeCategory)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-serif text-2xl font-bold text-wine-900">Menu Management</h1>
          <p className="text-wood-500">Manage dishes and presentations</p>
        </div>
        <button className="btn-elegant flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Add Dish
        </button>
      </div>
      
      {/* Category Tabs */}
      <div className="flex gap-2 border-b border-cream-200">
        {menuCategories.map((category) => (
          <button
            key={category.name}
            onClick={() => setActiveCategory(category.name)}
            className={cn(
              "px-6 py-3 font-medium transition-all border-b-2 -mb-px",
              activeCategory === category.name
                ? "border-wine-600 text-wine-900"
                : "border-transparent text-wood-500 hover:text-wood-700"
            )}
          >
            {category.name}
          </button>
        ))}
      </div>
      
      <div className="grid grid-cols-3 gap-6">
        {/* Menu Items */}
        <div className="col-span-2 space-y-4">
          {currentCategory?.items.map((item) => (
            <motion.div
              key={item.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              onClick={() => setSelectedItem(item)}
              className={cn(
                "elegant-card p-5 cursor-pointer transition-all hover:shadow-lg",
                selectedItem?.id === item.id && "ring-2 ring-wine-500"
              )}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <h3 className="font-serif text-lg font-semibold text-wine-900">
                      {item.name}
                    </h3>
                    {item.dietary?.includes('vegetarian') && (
                      <Leaf className="w-4 h-4 text-emerald-600" />
                    )}
                    {'winePairing' in item && item.winePairing && (
                      <Wine className="w-4 h-4 text-wine-500" />
                    )}
                  </div>
                  <p className="text-wood-600 mt-1">{item.description}</p>
                  
                  {/* Tags */}
                  <div className="flex flex-wrap gap-2 mt-3">
                    {item.dietary?.map((d) => (
                      <span key={d} className="px-2 py-0.5 bg-emerald-100 text-emerald-700 text-xs rounded-full capitalize">
                        {d}
                      </span>
                    ))}
                    {item.allergens?.map((a) => (
                      <span key={a} className="px-2 py-0.5 bg-amber-100 text-amber-700 text-xs rounded-full capitalize">
                        {a}
                      </span>
                    ))}
                  </div>
                </div>
                
                <div className="text-right">
                  <span className="font-serif text-2xl font-bold text-wine-900">
                    ${item.price}
                  </span>
                  {'winePairing' in item && item.winePairing && (
                    <p className="text-xs text-wood-500 mt-1">
                      Pairs with {item.winePairing}
                    </p>
                  )}
                </div>
              </div>
            </motion.div>
          ))}
        </div>
        
        {/* Item Details / Presentation Editor */}
        <div>
          {selectedItem ? (
            <motion.div
              key={selectedItem.id}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              className="elegant-card p-6 sticky top-24"
            >
              <h3 className="font-serif text-xl font-semibold text-wine-900 mb-4">
                Presentation Script
              </h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-wood-700 mb-1">
                    Opening
                  </label>
                  <textarea
                    className="w-full p-3 border border-cream-300 rounded-lg text-sm"
                    rows={2}
                    placeholder="Set the scene..."
                    defaultValue="From Chef Marco's autumn collection..."
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-wood-700 mb-1">
                    Story
                  </label>
                  <textarea
                    className="w-full p-3 border border-cream-300 rounded-lg text-sm"
                    rows={4}
                    placeholder="Tell the ingredient journey..."
                    defaultValue={`Hand-foraged ingredients from the Italian countryside, prepared with techniques passed down through generations...`}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-wood-700 mb-1">
                    Wine Pairing Note
                  </label>
                  <textarea
                    className="w-full p-3 border border-cream-300 rounded-lg text-sm"
                    rows={2}
                    placeholder="Isabella's suggestion..."
                    defaultValue={'winePairing' in selectedItem && selectedItem.winePairing ? `Isabella suggests the ${selectedItem.winePairing}, whose notes beautifully complement...` : ''}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-wood-700 mb-1">
                    Serving Instruction
                  </label>
                  <textarea
                    className="w-full p-3 border border-cream-300 rounded-lg text-sm"
                    rows={2}
                    placeholder="How to enjoy..."
                    defaultValue="Best enjoyed while warm, starting from the center..."
                  />
                </div>
                
                <button className="w-full btn-elegant">
                  Save Presentation
                </button>
              </div>
            </motion.div>
          ) : (
            <div className="elegant-card p-6 text-center text-wood-400">
              <Edit className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>Select a dish to edit its presentation</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
