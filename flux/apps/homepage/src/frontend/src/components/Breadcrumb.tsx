import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Home, ChevronRight } from 'lucide-react'

// =============================================================================
// 📋 TYPES
// =============================================================================

interface BreadcrumbItem {
  label: string
  path: string
  isCurrentPage: boolean
}

// =============================================================================
// 🗺️ ROUTE LABELS MAPPING
// =============================================================================

const routeLabels: Record<string, string> = {
  '': 'Home',
  'blog': 'Blog',
  'resume': 'Resume',
}

// =============================================================================
// 🎯 BREADCRUMB COMPONENT
// =============================================================================

const Breadcrumb: React.FC = () => {
  const location = useLocation()
  
  // Parse current path into breadcrumb items
  const getBreadcrumbItems = (): BreadcrumbItem[] => {
    const pathSegments = location.pathname.split('/').filter(Boolean)
    const items: BreadcrumbItem[] = []
    
    // Always add Home as the first item
    items.push({
      label: 'Home',
      path: '/',
      isCurrentPage: pathSegments.length === 0,
    })
    
    // Build breadcrumb path progressively
    let currentPath = ''
    pathSegments.forEach((segment, index) => {
      currentPath += `/${segment}`
      const isLast = index === pathSegments.length - 1
      
      // Get label from mapping or capitalize segment
      const label = routeLabels[segment] || 
        segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' ')
      
      items.push({
        label,
        path: currentPath,
        isCurrentPage: isLast,
      })
    })
    
    return items
  }

  const breadcrumbItems = getBreadcrumbItems()
  
  // Don't render breadcrumb on home page (only one item)
  if (breadcrumbItems.length <= 1) {
    return null
  }

  return (
    <nav className="breadcrumb-nav" aria-label="Breadcrumb">
      <ol className="breadcrumb-list">
        {breadcrumbItems.map((item, index) => (
          <li key={item.path} className="breadcrumb-item">
            {index > 0 && (
              <ChevronRight className="breadcrumb-separator" aria-hidden="true" />
            )}
            {item.isCurrentPage ? (
              <span className="breadcrumb-current" aria-current="page">
                {index === 0 && <Home className="breadcrumb-home-icon" aria-hidden="true" />}
                <span className="breadcrumb-label">{item.label}</span>
              </span>
            ) : (
              <Link to={item.path} className="breadcrumb-link">
                {index === 0 && <Home className="breadcrumb-home-icon" aria-hidden="true" />}
                <span className="breadcrumb-label">{item.label}</span>
              </Link>
            )}
          </li>
        ))}
      </ol>
    </nav>
  )
}

export default Breadcrumb
