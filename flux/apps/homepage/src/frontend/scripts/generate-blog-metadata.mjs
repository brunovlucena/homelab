#!/usr/bin/env node
/**
 * Generates public/blog-posts/metadata.json from markdown files with frontmatter.
 * Run before build so the blog page can load post listing.
 */
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'
import matter from 'gray-matter'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const frontendRoot = path.resolve(__dirname, '..')
const blogPostsDir = path.join(frontendRoot, 'public', 'blog-posts')
const metadataPath = path.join(blogPostsDir, 'metadata.json')

function toISODate (value) {
  if (value == null) return null
  if (value instanceof Date) return value.toISOString()
  const str = String(value).trim()
  if (/^\d{4}-\d{2}-\d{2}T/.test(str)) return str
  if (/^\d{4}-\d{2}-\d{2}$/.test(str)) return `${str}T00:00:00.000Z`
  return str
}

function generate () {
  if (!fs.existsSync(blogPostsDir)) {
    fs.writeFileSync(metadataPath, '[]', 'utf8')
    return
  }

  const files = fs.readdirSync(blogPostsDir).filter(f => f.endsWith('.md'))
  const metadata = files.map((filename, index) => {
    const filePath = path.join(blogPostsDir, filename)
    const content = fs.readFileSync(filePath, 'utf8')
    const { data } = matter(content)

    const publishedDate = toISODate(data.publishedDate) ?? toISODate(data.date) ?? new Date().toISOString()
    const updatedDate = toISODate(data.updatedDate) ?? publishedDate

    return {
      id: String(data.id ?? index + 1),
      slug: data.slug ?? path.basename(filename, '.md'),
      title: data.title ?? path.basename(filename, '.md'),
      filename,
      authors: Array.isArray(data.authors) ? data.authors : (data.author ? [data.author] : []),
      tags: Array.isArray(data.tags) ? data.tags : [],
      publishedDate,
      updatedDate,
      readTime: data.readTime ?? '1 min read',
    }
  })

  // Sort by publishedDate descending
  metadata.sort((a, b) => new Date(b.publishedDate) - new Date(a.publishedDate))

  fs.writeFileSync(metadataPath, JSON.stringify(metadata, null, 2), 'utf8')
}

generate()
