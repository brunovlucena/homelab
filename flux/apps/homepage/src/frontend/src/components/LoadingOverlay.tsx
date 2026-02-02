import React from 'react'
import './LoadingOverlay.css'

interface LoadingOverlayProps {
  isLoading: boolean
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({ isLoading }) => {
  if (!isLoading) {
    return null;
  }

  return (
    <div className="loading-overlay">
      <div className="loading-spinner">
        <div className="spinner-ring"></div>
        <div className="spinner-ring"></div>
        <div className="spinner-ring"></div>
      </div>
    </div>
  )
}

export default LoadingOverlay
