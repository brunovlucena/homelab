import { useState, useEffect } from 'react'
import FileUpload from './components/FileUpload'
import FileList from './components/FileList'
import './App.css'

type Target = 'lambda' | 'agent'

interface Resource {
  name: string
  namespace: string
  status?: any
  labels?: Record<string, string>
}

const API_BASE = (import.meta.env?.VITE_API_BASE as string) || '/api/v1'

function App() {
  const [target, setTarget] = useState<Target>('agent')
  const [resources, setResources] = useState<Resource[]>([])
  const [selectedResource, setSelectedResource] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [refreshKey, setRefreshKey] = useState(0)

  const handleUploadComplete = () => {
    // Refresh file list after upload
    setRefreshKey(prev => prev + 1)
  }

  // Fetch resources from Kubernetes
  useEffect(() => {
    const fetchResources = async () => {
      setLoading(true)
      try {
        const endpoint = target === 'agent' ? '/agents' : '/lambdas'
        const response = await fetch(`${API_BASE}${endpoint}`)
        if (!response.ok) {
          throw new Error('Failed to fetch resources')
        }
        const data = await response.json()
        setResources(data.items || [])
        
        // Auto-select first resource if available
        if (data.items && data.items.length > 0 && !selectedResource) {
          setSelectedResource(data.items[0].name)
        }
      } catch (error) {
        console.error('Failed to fetch resources:', error)
        setResources([])
      } finally {
        setLoading(false)
      }
    }

    fetchResources()
  }, [target])

  // Reset selection when target changes
  useEffect(() => {
    if (resources.length > 0 && !resources.find(r => r.name === selectedResource)) {
      setSelectedResource(resources[0]?.name || '')
    }
  }, [resources, selectedResource])

  return (
    <div className="app">
      <header className="app-header">
        <h1>🤖 Agent File Manager</h1>
        <p className="subtitle">Manage Training Files, Code & Documentation for AI Agents</p>
      </header>

      <div className="target-selector">
        <button
          className={`target-btn ${target === 'lambda' ? 'active' : ''}`}
          onClick={() => setTarget('lambda')}
        >
          📦 Lambda Functions
        </button>
        <button
          className={`target-btn ${target === 'agent' ? 'active' : ''}`}
          onClick={() => setTarget('agent')}
        >
          🤖 Chatbot Agents
        </button>
      </div>

      {target === 'agent' && (
        <div className="agent-selector">
          <label htmlFor="resource-select">
            Select {target === 'agent' ? 'Agent' : 'Lambda Function'}:
          </label>
          {loading ? (
            <div className="loading">Loading from Kubernetes...</div>
          ) : resources.length === 0 ? (
            <div className="no-resources">
              No {target === 'agent' ? 'agents' : 'lambda functions'} found in Kubernetes
            </div>
          ) : (
            <>
              <select
                id="resource-select"
                value={selectedResource}
                onChange={(e) => setSelectedResource(e.target.value)}
                className="agent-select"
              >
                {resources.map((resource) => (
                  <option key={resource.name} value={resource.name}>
                    {resource.name} {resource.namespace ? `(${resource.namespace})` : ''}
                  </option>
                ))}
              </select>
              <div className="agent-info">
                📁 Files will be uploaded to:{' '}
                <code>
                  {target === 'agent' ? 'agent-files' : 'lambda-functions'}/{selectedResource}/
                </code>
              </div>
            </>
          )}
        </div>
      )}

      {target === 'lambda' && (
        <div className="agent-selector">
          <label htmlFor="resource-select">Select Lambda Function:</label>
          {loading ? (
            <div className="loading">Loading from Kubernetes...</div>
          ) : resources.length === 0 ? (
            <div className="no-resources">No lambda functions found in Kubernetes</div>
          ) : (
            <>
              <select
                id="resource-select"
                value={selectedResource}
                onChange={(e) => setSelectedResource(e.target.value)}
                className="agent-select"
              >
                {resources.map((resource) => (
                  <option key={resource.name} value={resource.name}>
                    {resource.name} {resource.namespace ? `(${resource.namespace})` : ''}
                  </option>
                ))}
              </select>
              <div className="agent-info">
                📁 Files will be uploaded to: <code>lambda-functions/{selectedResource}/</code>
              </div>
            </>
          )}
        </div>
      )}

      <div className="content">
        <div className="upload-section">
          <h2>
            {target === 'agent' ? '📚 Upload Learning Materials' : 'Upload Files'}
          </h2>
          {target === 'agent' && (
            <p className="section-description">
              Upload runbooks, documentation, and training materials for fine-tuning.
              Files will be organized in the agent's folder.
            </p>
          )}
          <FileUpload
            target={target}
            agentPath={selectedResource || undefined}
            onUploadComplete={handleUploadComplete}
          />
        </div>

        <div className="files-section">
          <h2>
            {target === 'agent' ? '📋 File Library' : '📦 File Library'}
          </h2>
          {target === 'agent' && selectedResource && (
            <p className="section-description">
              View, edit, and manage all files for <strong>{selectedResource}</strong> • Switch between List and Tree view
            </p>
          )}
          {target === 'lambda' && selectedResource && (
            <p className="section-description">
              View, edit, and manage all files for <strong>{selectedResource}</strong> • Switch between List and Tree view
            </p>
          )}
          <FileList
            target={target}
            prefix={selectedResource ? `${selectedResource}/` : undefined}
            key={`${refreshKey}-${selectedResource}`}
          />
        </div>
      </div>
    </div>
  )
}

export default App
