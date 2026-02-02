import { useState, useEffect } from 'react'
import { X, Loader, FileText, Code, Image, File } from 'lucide-react'
import './FileViewer.css'

interface FileViewerProps {
  target: 'lambda' | 'agent'
  filePath: string
  fileName: string
  onClose: () => void
  onEdit?: () => void
}

const API_BASE = (import.meta.env?.VITE_API_BASE as string) || '/api/v1'

export default function FileViewer({ target, filePath, fileName, onClose, onEdit }: FileViewerProps) {
  const [content, setContent] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isText, setIsText] = useState(true)

  useEffect(() => {
    fetchFile()
  }, [filePath])

  const fetchFile = async () => {
    setLoading(true)
    setError(null)
    try {
      const encodedPath = encodeURIComponent(filePath)
      const response = await fetch(`${API_BASE}/files/${target}/${encodedPath}`)
      if (!response.ok) {
        throw new Error('Failed to fetch file')
      }
      const data = await response.json()
      setContent(data.content || '')
      setIsText(data.isText || false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load file')
    } finally {
      setLoading(false)
    }
  }

  const getFileIcon = () => {
    const ext = fileName.split('.').pop()?.toLowerCase()
    if (['js', 'ts', 'jsx', 'tsx', 'py', 'java', 'go', 'rs', 'cpp', 'c'].includes(ext || '')) {
      return <Code size={20} />
    }
    if (['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp'].includes(ext || '')) {
      return <Image size={20} />
    }
    return <FileText size={20} />
  }

  const getLanguage = () => {
    const ext = fileName.split('.').pop()?.toLowerCase()
    const langMap: Record<string, string> = {
      'js': 'javascript',
      'jsx': 'javascript',
      'ts': 'typescript',
      'tsx': 'typescript',
      'py': 'python',
      'java': 'java',
      'go': 'go',
      'rs': 'rust',
      'cpp': 'cpp',
      'c': 'c',
      'json': 'json',
      'yaml': 'yaml',
      'yml': 'yaml',
      'html': 'html',
      'css': 'css',
      'md': 'markdown',
      'sh': 'bash',
      'xml': 'xml',
    }
    return langMap[ext || ''] || 'plaintext'
  }

  if (loading) {
    return (
      <div className="file-viewer-overlay" onClick={onClose}>
        <div className="file-viewer-modal" onClick={(e) => e.stopPropagation()}>
          <div className="file-viewer-loading">
            <Loader className="spinner" size={32} />
            <p>Loading file...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="file-viewer-overlay" onClick={onClose}>
        <div className="file-viewer-modal" onClick={(e) => e.stopPropagation()}>
          <div className="file-viewer-error">
            <p>Error: {error}</p>
            <button onClick={onClose}>Close</button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="file-viewer-overlay" onClick={onClose}>
      <div className="file-viewer-modal" onClick={(e) => e.stopPropagation()}>
        <div className="file-viewer-header">
          <div className="file-viewer-title">
            {getFileIcon()}
            <span>{fileName}</span>
          </div>
          <div className="file-viewer-actions">
            {isText && onEdit && (
              <button className="edit-btn" onClick={onEdit} title="Edit file">
                Edit
              </button>
            )}
            <button className="close-btn" onClick={onClose} title="Close">
              <X size={20} />
            </button>
          </div>
        </div>
        <div className="file-viewer-content">
          {isText ? (
            <pre className="file-content">
              <code className={`language-${getLanguage()}`}>{content}</code>
            </pre>
          ) : (
            <div className="file-content-binary">
              <p>Binary file detected. Size: {(content.length / 1024).toFixed(2)} KB</p>
              <p className="binary-hint">Binary files cannot be displayed in the viewer.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
