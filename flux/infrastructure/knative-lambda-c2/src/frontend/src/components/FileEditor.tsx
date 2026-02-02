import { useState, useEffect } from 'react'
import { X, Save, Loader, AlertCircle } from 'lucide-react'
import './FileEditor.css'

interface FileEditorProps {
  target: 'lambda' | 'agent'
  filePath: string
  fileName: string
  initialContent: string
  onClose: () => void
  onSave?: () => void
}

const API_BASE = (import.meta.env?.VITE_API_BASE as string) || '/api/v1'

export default function FileEditor({ target, filePath, fileName, initialContent, onClose, onSave }: FileEditorProps) {
  const [content, setContent] = useState(initialContent)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasChanges, setHasChanges] = useState(false)

  useEffect(() => {
    setContent(initialContent)
    setHasChanges(false)
  }, [initialContent])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setContent(e.target.value)
    setHasChanges(true)
  }

  const handleSave = async () => {
    setSaving(true)
    setError(null)
    try {
      const encodedPath = encodeURIComponent(filePath)
      const response = await fetch(`${API_BASE}/files/${target}/${encodedPath}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          content,
          isText: true,
          contentType: 'text/plain'
        }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.detail || 'Failed to save file')
      }

      setHasChanges(false)
      if (onSave) {
        onSave()
      }
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save file')
    } finally {
      setSaving(false)
    }
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

  return (
    <div className="file-editor-overlay" onClick={onClose}>
      <div className="file-editor-modal" onClick={(e) => e.stopPropagation()}>
        <div className="file-editor-header">
          <div className="file-editor-title">
            <span>Editing: {fileName}</span>
          </div>
          <div className="file-editor-actions">
            <button
              className="save-btn"
              onClick={handleSave}
              disabled={saving || !hasChanges}
              title="Save changes"
            >
              {saving ? <Loader className="spinner" size={16} /> : <Save size={16} />}
              Save
            </button>
            <button className="close-btn" onClick={onClose} title="Close">
              <X size={20} />
            </button>
          </div>
        </div>

        {error && (
          <div className="file-editor-error">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        <div className="file-editor-content">
          <textarea
            className={`file-editor-textarea language-${getLanguage()}`}
            value={content}
            onChange={handleChange}
            spellCheck={false}
            placeholder="File content..."
          />
        </div>

        {hasChanges && !saving && (
          <div className="file-editor-footer">
            <span className="unsaved-indicator">● Unsaved changes</span>
          </div>
        )}
      </div>
    </div>
  )
}
