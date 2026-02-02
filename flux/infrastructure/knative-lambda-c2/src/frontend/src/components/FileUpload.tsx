import { useState, useEffect } from 'react'
import { useDropzone } from 'react-dropzone'
import { Upload, X, Check, AlertCircle, Loader } from 'lucide-react'
import './FileUpload.css'

interface FileUploadProps {
  target: 'lambda' | 'agent'
  agentPath?: string
  onUploadComplete: () => void
}

interface FileState {
  file: File
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
  error?: string
  fileId?: string
}

const API_BASE = (import.meta.env?.VITE_API_BASE as string) || '/api/v1'

export default function FileUpload({ target, agentPath, onUploadComplete }: FileUploadProps) {
  const [files, setFiles] = useState<FileState[]>([])
  const [path, setPath] = useState(agentPath || '')
  
  // Update path when agent changes
  useEffect(() => {
    if (agentPath) {
      setPath(agentPath)
    }
  }, [agentPath])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop: async (acceptedFiles) => {
      for (const file of acceptedFiles) {
        await uploadFile(file)
      }
    },
    maxSize: 100 * 1024 * 1024, // 100 MB
  })

  async function uploadFile(file: File) {
    const fileState: FileState = {
      file,
      progress: 0,
      status: 'pending',
    }
    setFiles(prev => [...prev, fileState])

    try {
      // Step 1: Request presigned URL
      fileState.status = 'uploading'
      fileState.progress = 10
      setFiles(prev => [...prev])

      const presignedResponse = await fetch(`${API_BASE}/files/presigned-url`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          filename: file.name,
          mimeType: file.type || 'application/octet-stream',
          size: file.size,
          target: target,
          path: path ? `${path}/` : undefined,
        }),
      })

      if (!presignedResponse.ok) {
        const error = await presignedResponse.json()
        throw new Error(error.detail || 'Failed to get presigned URL')
      }

      const { uploadUrl, fileId, objectPath } = await presignedResponse.json()
      fileState.fileId = fileId
      fileState.progress = 30
      setFiles(prev => [...prev])

      // Step 2: Upload directly to MinIO
      const uploadResponse = await fetch(uploadUrl, {
        method: 'PUT',
        body: file,
        headers: {
          'Content-Type': file.type || 'application/octet-stream',
        },
      })

      if (!uploadResponse.ok) {
        throw new Error('Upload to MinIO failed')
      }

      fileState.progress = 90
      setFiles(prev => [...prev])

      // Step 3: Notify backend upload complete
      await fetch(`${API_BASE}/files/${fileId}/complete`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fileId,
          objectPath,
        }),
      })

      fileState.status = 'success'
      fileState.progress = 100
      setFiles(prev => [...prev])

      // Call callback after a short delay
      setTimeout(() => {
        onUploadComplete()
      }, 1000)
    } catch (error) {
      fileState.status = 'error'
      fileState.error = error instanceof Error ? error.message : 'Upload failed'
      setFiles(prev => [...prev])
    }
  }

  function removeFile(index: number) {
    setFiles(prev => prev.filter((_, i) => i !== index))
  }

  return (
    <div className="file-upload">
      {target === 'agent' ? (
        <div className="path-display">
          <label>Agent Folder:</label>
          <div className="path-value">
            <code>{path}/</code>
          </div>
          <p className="path-hint">Files will be uploaded to this agent's folder</p>
        </div>
      ) : (
        <div className="path-input">
          <label htmlFor="path">Optional Path (e.g., "my-function/"):</label>
          <input
            id="path"
            type="text"
            value={path}
            onChange={(e) => setPath(e.target.value)}
            placeholder="Leave empty for default structure"
          />
        </div>
      )}

      <div
        {...getRootProps()}
        className={`dropzone ${isDragActive ? 'active' : ''}`}
      >
        <input {...getInputProps()} />
        <Upload size={48} />
        <p>
          {isDragActive
            ? 'Drop files here...'
            : 'Drag & drop files here, or click to select'}
        </p>
        <p className="hint">Max file size: 100 MB</p>
      </div>

      {files.length > 0 && (
        <div className="file-list">
          {files.map((fileState, index) => (
            <div key={index} className={`file-item ${fileState.status}`}>
              <div className="file-info">
                <div className="file-name">{fileState.file.name}</div>
                <div className="file-size">
                  {(fileState.file.size / 1024 / 1024).toFixed(2)} MB
                </div>
              </div>

              <div className="file-status">
                {fileState.status === 'pending' && (
                  <Loader className="spinner" size={20} />
                )}
                {fileState.status === 'uploading' && (
                  <>
                    <Loader className="spinner" size={20} />
                    <div className="progress-bar">
                      <div
                        className="progress-fill"
                        style={{ width: `${fileState.progress}%` }}
                      />
                    </div>
                    <span>{fileState.progress}%</span>
                  </>
                )}
                {fileState.status === 'success' && (
                  <Check size={20} className="success-icon" />
                )}
                {fileState.status === 'error' && (
                  <AlertCircle size={20} className="error-icon" />
                )}
              </div>

              {fileState.error && (
                <div className="error-message">{fileState.error}</div>
              )}

              {fileState.status !== 'uploading' && (
                <button
                  className="remove-btn"
                  onClick={() => removeFile(index)}
                  aria-label="Remove file"
                >
                  <X size={16} />
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
