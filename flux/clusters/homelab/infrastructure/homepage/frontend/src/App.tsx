import React, { useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { Toaster } from 'react-hot-toast'
import Header from './components/Header'
import Home from './pages/Home'
import Resume from './pages/Resume'
import Chatbot from './components/Chatbot'
import { ChatbotProvider } from './contexts/ChatbotContext'
import { usePageViewTracking, useSessionTracking } from './hooks/useMetrics'
import { initWebVitals } from './utils/webVitals'
import { metricsCollector } from './utils/metrics'
import './App.css'

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
})

// 📊 Metrics-aware App component
function AppContent() {
  // Track page views automatically
  usePageViewTracking()
  
  // Track session duration
  useSessionTracking()

  return (
    <div className="App">
      <Header />
      <main>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/resume" element={<Resume />} />
        </Routes>
      </main>
      <Chatbot />
      <Toaster 
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#363636',
            color: '#fff',
          },
        }}
      />
    </div>
  )
}

function App() {
  useEffect(() => {
    // 📊 Initialize Web Vitals tracking on mount
    initWebVitals()
    
    // 📊 Track navigation timing
    if (typeof window !== 'undefined') {
      // Wait for page to fully load before tracking timing
      window.addEventListener('load', () => {
        setTimeout(() => {
          metricsCollector.recordNavigationTiming()
          metricsCollector.recordResourceTiming()
        }, 0)
      })
    }

    // 📊 Track global errors
    const handleError = (event: ErrorEvent) => {
      metricsCollector.recordError(
        'uncaught_error',
        event.message,
        event.error?.stack
      )
    }

    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      metricsCollector.recordError(
        'unhandled_promise_rejection',
        event.reason?.message || String(event.reason),
        event.reason?.stack
      )
    }

    window.addEventListener('error', handleError)
    window.addEventListener('unhandledrejection', handleUnhandledRejection)

    return () => {
      window.removeEventListener('error', handleError)
      window.removeEventListener('unhandledrejection', handleUnhandledRejection)
    }
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
      <ChatbotProvider>
        <Router>
          <AppContent />
        </Router>
        <ReactQueryDevtools initialIsOpen={false} />
      </ChatbotProvider>
    </QueryClientProvider>
  )
}

export default App
