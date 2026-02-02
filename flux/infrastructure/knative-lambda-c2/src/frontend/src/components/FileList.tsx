import { useState, useEffect } from 'react'
import { Trash2, RefreshCw, Loader, Eye, Edit2, Copy, Download, List, FolderTree } from 'lucide-react'
import FileViewer from './FileViewer'
import FileEditor from './FileEditor'
import FileTree from './FileTree'
import './FileList.css'

interface FileListProps {
  target: 'lambda' | 'agent'
  prefix?: string
}

interface File {
  name: string
  size: number
  lastModified: string | null
  etag: string
}

const API_BASE = (import.meta.env?.VITE_API_BASE as string) || '/api/v1'

export default function FileList({ target, prefix }: FileListProps) {
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [viewingFile, setViewingFile] = useState<string | null>(null)
  const [editingFile, setEditingFile] = useState<string | null>(null)
  const [editingContent, setEditingContent] = useState<string>('')
  const [viewMode, setViewMode] = useState<'list' | 'tree'>('list')

  const fetchFiles = async () => {
    setLoading(true)
    setError(null)
    try {
      const url = new URL(`${API_BASE}/files/list`, window.location.origin)
      url.searchParams.set('target', target)
      if (prefix) {
        url.searchParams.set('prefix', prefix)
      }
      const response = await fetch(url.toString())
      if (!response.ok) {
        throw new Error('Failed to fetch files')
      }
      const data = await response.json()
      setFiles(data.files || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load files')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchFiles()
  }, [target, prefix])

  const handleView = async (file: File) => {
    setViewingFile(file.name)
  }

  const handleEdit = async (file: File) => {
    try {
      const encodedPath = encodeURIComponent(file.name)
      const response = await fetch(`${API_BASE}/files/${target}/${encodedPath}`)
      if (!response.ok) {
        throw new Error('Failed to fetch file for editing')
      }
      const data = await response.json()
      if (!data.isText) {
        alert('Only text files can be edited')
        return
      }
      setEditingContent(data.content)
      setEditingFile(file.name)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to load file for editing')
    }
  }

  const handleCopy = async (file: File) => {
    const fileName = file.name.split('/').pop() || file.name
    const newName = prompt('Enter new filename:', fileName)
    if (!newName || newName === fileName) {
      return
    }

    try {
      // Construct destination path maintaining the same folder structure
      const pathParts = file.name.split('/')
      pathParts[pathParts.length - 1] = newName
      const destinationPath = pathParts.join('/')

      const response = await fetch(`${API_BASE}/files/${target}/${encodeURIComponent(file.name)}/copy`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          destinationPath: destinationPath
        }),
      })
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.detail || 'Failed to copy file')
      }
      await fetchFiles()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to copy file')
    }
  }

  const handleDownload = async (file: File) => {
    try {
      const encodedPath = encodeURIComponent(file.name)
      const response = await fetch(`${API_BASE}/files/${target}/${encodedPath}`)
      if (!response.ok) {
        throw new Error('Failed to download file')
      }
      const data = await response.json()
      
      let blob: Blob
      if (data.isText) {
        blob = new Blob([data.content], { type: 'text/plain' })
      } else {
        // Binary file - decode base64
        const binaryString = atob(data.content)
        const bytes = new Uint8Array(binaryString.length)
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i)
        }
        blob = new Blob([bytes])
      }
      
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.name.split('/').pop() || 'download'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to download file')
    }
  }

  const handleDelete = async (path: string) => {
    if (!confirm(`Are you sure you want to delete ${path}?`)) {
      return
    }

    try {
      const encodedPath = encodeURIComponent(path)
      const response = await fetch(`${API_BASE}/files/${target}/${encodedPath}`, {
        method: 'DELETE',
      })
      if (!response.ok) {
        throw new Error('Failed to delete file')
      }
      await fetchFiles()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete file')
    }
  }

  const isTextFile = (fileName: string): boolean => {
    const ext = fileName.split('.').pop()?.toLowerCase()
    const textExtensions = ['js', 'ts', 'jsx', 'tsx', 'py', 'java', 'go', 'rs', 'cpp', 'c', 'json', 'yaml', 'yml', 'html', 'css', 'md', 'sh', 'txt', 'xml', 'yml', 'yaml', 'env', 'conf', 'config']
    return textExtensions.includes(ext || '')
  }

  const formatSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  }

  const formatDate = (dateString: string | null): string => {
    if (!dateString) return 'Unknown'
    try {
      return new Date(dateString).toLocaleString()
    } catch {
      return 'Unknown'
    }
  }

  if (loading) {
    return (
      <div className="file-list-loading">
        <Loader className="spinner" size={24} />
        <p>Loading files...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="file-list-error">
        <p>Error: {error}</p>
        <button onClick={fetchFiles} className="retry-btn">
          <RefreshCw size={16} />
          Retry
        </button>
      </div>
    )
  }

  const handleFileSelect = (path: string) => {
    const file = files.find(f => f.name === path)
    if (file) {
      handleView(file)
    }
  }

  return (
    <div className="file-list">
      <div className="file-list-header">
        <span>{files.length} file{files.length !== 1 ? 's' : ''}</span>
        <div className="file-list-controls">
          <div className="view-mode-toggle">
            <button
              className={`view-mode-btn ${viewMode === 'list' ? 'active' : ''}`}
              onClick={() => setViewMode('list')}
              title="List view"
            >
              <List size={16} />
            </button>
            <button
              className={`view-mode-btn ${viewMode === 'tree' ? 'active' : ''}`}
              onClick={() => setViewMode('tree')}
              title="Tree view"
            >
              <FolderTree size={16} />
            </button>
          </div>
          <button onClick={fetchFiles} className="refresh-btn" title="Refresh">
            <RefreshCw size={16} />
          </button>
        </div>
      </div>

      {files.length === 0 ? (
        <div className="empty-state">
          <p>
            {prefix
              ? `No learning materials uploaded for this agent yet`
              : 'No files uploaded yet'}
          </p>
          {prefix && (
            <p className="empty-hint">
              Upload runbooks, documentation, or training materials to get started
            </p>
          )}
        </div>
      ) : viewMode === 'tree' ? (
        <FileTree
          files={files}
          onFileSelect={handleFileSelect}
          selectedFile={viewingFile}
        />
      ) : (
        <div className="file-items">
          {files.map((file, index) => {
            const displayName = prefix ? file.name.replace(prefix, '') : file.name
            const canEdit = isTextFile(file.name)
            
            return (
              <div key={index} className="file-item">
                <div className="file-details">
                  <div className="file-name" title={file.name}>
                    {displayName}
                  </div>
                  <div className="file-meta">
                    <span>{formatSize(file.size)}</span>
                    {file.lastModified && (
                      <>
                        <span>•</span>
                        <span>{formatDate(file.lastModified)}</span>
                      </>
                    )}
                  </div>
                </div>
                <div className="file-actions">
                  <button
                    className="action-btn view-btn"
                    onClick={() => handleView(file)}
                    title="View file"
                  >
                    <Eye size={16} />
                  </button>
                  {canEdit && (
                    <button
                      className="action-btn edit-btn"
                      onClick={() => handleEdit(file)}
                      title="Edit file"
                    >
                      <Edit2 size={16} />
                    </button>
                  )}
                  <button
                    className="action-btn copy-btn"
                    onClick={() => handleCopy(file)}
                    title="Copy file"
                  >
                    <Copy size={16} />
                  </button>
                  <button
                    className="action-btn download-btn"
                    onClick={() => handleDownload(file)}
                    title="Download file"
                  >
                    <Download size={16} />
                  </button>
                  <button
                    className="action-btn delete-btn"
                    onClick={() => handleDelete(file.name)}
                    title="Delete file"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {viewingFile && (
        <FileViewer
          target={target}
          filePath={viewingFile}
          fileName={viewingFile.split('/').pop() || viewingFile}
          onClose={() => setViewingFile(null)}
          onEdit={() => {
            setViewingFile(null)
            handleEdit(files.find(f => f.name === viewingFile)!)
          }}
        />
      )}

      {editingFile && (
        <FileEditor
          target={target}
          filePath={editingFile}
          fileName={editingFile.split('/').pop() || editingFile}
          initialContent={editingContent}
          onClose={() => {
            setEditingFile(null)
            setEditingContent('')
          }}
          onSave={() => {
            fetchFiles()
          }}
        />
      )}
    </div>
  )
}
