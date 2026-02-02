import { useState } from 'react'
import { ChevronRight, ChevronDown, Folder, File as FileIcon } from 'lucide-react'
import './FileTree.css'

interface FileTreeProps {
  files: Array<{
    name: string
    size: number
    lastModified: string | null
  }>
  onFileSelect: (path: string) => void
  selectedFile?: string | null
}

interface TreeNode {
  name: string
  fullPath: string
  children: Map<string, TreeNode>
  files: Array<{
    name: string
    fullPath: string
    size: number
    lastModified: string | null
  }>
  isExpanded: boolean
}

function buildTree(files: Array<{ name: string; size: number; lastModified: string | null }>): TreeNode {
  const root: TreeNode = {
    name: '',
    fullPath: '',
    children: new Map(),
    files: [],
    isExpanded: true,
  }

  for (const file of files) {
    const parts = file.name.split('/').filter(Boolean)
    let current = root

    for (let i = 0; i < parts.length - 1; i++) {
      const part = parts[i]
      if (!current.children.has(part)) {
        current.children.set(part, {
          name: part,
          fullPath: parts.slice(0, i + 1).join('/'),
          children: new Map(),
          files: [],
          isExpanded: false,
        })
      }
      current = current.children.get(part)!
    }

    const fileName = parts[parts.length - 1]
    current.files.push({
      name: fileName,
      fullPath: file.name,
      size: file.size,
      lastModified: file.lastModified,
    })
  }

  return root
}

interface TreeNodeProps {
  node: TreeNode
  level: number
  onFileSelect: (path: string) => void
  selectedFile?: string | null
}

function TreeNodeComponent({ node, level, onFileSelect, selectedFile }: TreeNodeProps) {
  const [isExpanded, setIsExpanded] = useState(node.isExpanded)

  const toggleExpanded = () => {
    setIsExpanded(!isExpanded)
  }

  const sortedFolders = Array.from(node.children.entries()).sort(([a], [b]) => a.localeCompare(b))
  const sortedFiles = [...node.files].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <div className="tree-node">
      {node.name && (
        <div
          className={`tree-folder ${isExpanded ? 'expanded' : ''}`}
          style={{ paddingLeft: `${level * 16}px` }}
          onClick={toggleExpanded}
        >
          {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
          <Folder size={16} />
          <span className="folder-name">{node.name}</span>
        </div>
      )}

      {isExpanded && (
        <>
          {sortedFolders.map(([folderName, childNode]) => (
            <TreeNodeComponent
              key={childNode.fullPath}
              node={childNode}
              level={level + 1}
              onFileSelect={onFileSelect}
              selectedFile={selectedFile}
            />
          ))}

          {sortedFiles.map((file) => (
            <div
              key={file.fullPath}
              className={`tree-file ${selectedFile === file.fullPath ? 'selected' : ''}`}
              style={{ paddingLeft: `${(level + 1) * 16}px` }}
              onClick={() => onFileSelect(file.fullPath)}
            >
              <FileIcon size={16} />
              <span className="file-name">{file.name}</span>
              <span className="file-size">{(file.size / 1024).toFixed(1)} KB</span>
            </div>
          ))}
        </>
      )}
    </div>
  )
}

export default function FileTree({ files, onFileSelect, selectedFile }: FileTreeProps) {
  const tree = buildTree(files)

  if (files.length === 0) {
    return (
      <div className="file-tree-empty">
        <p>No files to display</p>
      </div>
    )
  }

  return (
    <div className="file-tree">
      <TreeNodeComponent
        node={tree}
        level={0}
        onFileSelect={onFileSelect}
        selectedFile={selectedFile}
      />
    </div>
  )
}
