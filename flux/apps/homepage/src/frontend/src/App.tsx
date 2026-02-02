import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { Toaster } from 'react-hot-toast'
import Header from './components/Header'
import Home from './pages/Home'
import Resume from './pages/Resume'
import Blog from './pages/Blog'
import Chatbot from './components/Chatbot'
import { ChatbotProvider } from './contexts/ChatbotContext'
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

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ChatbotProvider>
        <Router>
          <div className="App">
            <Header />
            <main>
              <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/resume" element={<Resume />} />
                <Route path="/blog" element={<Blog />} />
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
        </Router>
        <ReactQueryDevtools initialIsOpen={false} />
      </ChatbotProvider>
    </QueryClientProvider>
  )
}

export default App
