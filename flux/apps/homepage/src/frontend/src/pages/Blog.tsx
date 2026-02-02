import React, { useState, useEffect, useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import rehypeSanitize from 'rehype-sanitize'
import matter from 'gray-matter'
import { useInView } from 'react-intersection-observer'

// Constants
const BLOG_POSTS_BASE_URL = '/blog-posts/'
const METADATA_FILE = 'metadata.json'
const POSTS_PER_PAGE = 5
const CACHE_DURATION_MS = 5 * 60 * 1000 // 5 minutes

interface BlogPostMetadata {
  id: string
  slug: string
  title: string
  filename: string
  authors: string[]
  tags: string[]
  publishedDate: string
  updatedDate: string
  readTime: string
  published?: boolean // Optional field to control visibility (default: true)
}

interface BlogPost extends BlogPostMetadata {
  content: string
  isLoaded: boolean
}

interface CacheEntry {
  data: BlogPostMetadata[]
  timestamp: number
}

const Blog: React.FC = () => {
  const [blogPosts, setBlogPosts] = useState<BlogPost[]>([])
  const [displayedPosts, setDisplayedPosts] = useState<BlogPost[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(true)

  // Infinite scroll trigger
  const { ref: loadMoreRef, inView } = useInView({
    threshold: 0,
    triggerOnce: false,
  })

  // Cache management
  const getCachedMetadata = (): BlogPostMetadata[] | null => {
    try {
      const cached = localStorage.getItem('blog_metadata_cache')
      if (!cached) {
        return null;
      }

      const cacheEntry: CacheEntry = JSON.parse(cached)
      const now = Date.now()

      // Check if cache is still valid
      if (now - cacheEntry.timestamp < CACHE_DURATION_MS) {
        return cacheEntry.data
      }

      // Cache expired
      localStorage.removeItem('blog_metadata_cache')
      return null
    } catch {
      return null
    }
  }

  const setCachedMetadata = (data: BlogPostMetadata[]) => {
    try {
      const cacheEntry: CacheEntry = {
        data,
        timestamp: Date.now(),
      }
      localStorage.setItem('blog_metadata_cache', JSON.stringify(cacheEntry))
    } catch {
      // Silently fail if localStorage is not available
    }
  }

  // Fetch metadata only
  const fetchMetadata = async (): Promise<BlogPostMetadata[]> => {
    // Check cache first
    const cached = getCachedMetadata()
    if (cached) {
      return cached
    }

    const metadataResponse = await fetch(`${BLOG_POSTS_BASE_URL}${METADATA_FILE}`)
    if (!metadataResponse.ok) {
      throw new Error('Failed to fetch blog metadata')
    }
    const metadata: BlogPostMetadata[] = await metadataResponse.json()

    // Filter out unpublished posts
    const publishedMetadata = metadata.filter(meta => meta.published !== false)

    // Cache the metadata
    setCachedMetadata(publishedMetadata)

    return publishedMetadata
  }

  // Lazy load content for a specific post
  const loadPostContent = async (post: BlogPost): Promise<BlogPost> => {
    if (post.isLoaded) {
      return post;
    }

    try {
      const contentResponse = await fetch(`${BLOG_POSTS_BASE_URL}${post.filename}`)
      if (!contentResponse.ok) {
        throw new Error(`Failed to fetch ${post.filename}`)
      }
      const rawContent = await contentResponse.text()
      const { content } = matter(rawContent)

      return { ...post, content, isLoaded: true }
    } catch (err) {
      console.error(`Error loading content for ${post.filename}:`, err)
      return { ...post, content: 'Failed to load content', isLoaded: false }
    }
  }

  // Initial load
  useEffect(() => {
    const initializeBlog = async () => {
      try {
        const metadata = await fetchMetadata()

        // Sort by published date (newest first)
        metadata.sort((a, b) => 
          new Date(b.publishedDate).getTime() - new Date(a.publishedDate).getTime()
        )

        // Initialize posts without content
        const initialPosts: BlogPost[] = metadata.map(meta => ({
          ...meta,
          content: '',
          isLoaded: false,
        }))

        setBlogPosts(initialPosts)
        setLoading(false)
      } catch (err) {
        console.error('Error loading blog posts:', err)
        setError(err instanceof Error ? err.message : 'Failed to load blog posts')
        setLoading(false)
      }
    }

    void initializeBlog()
  }, [])

  // Load more posts when scrolling
  const loadMorePosts = useCallback(() => {
    if (!hasMore || loading) {
      return;
    }

    const startIdx = (currentPage - 1) * POSTS_PER_PAGE
    const endIdx = startIdx + POSTS_PER_PAGE
    const postsToLoad = blogPosts.slice(startIdx, endIdx)

    if (postsToLoad.length === 0) {
      setHasMore(false)
      return
    }

    // Load content for these posts
    Promise.all(postsToLoad.map(loadPostContent))
      .then(loadedPosts => {
        setDisplayedPosts(prev => [...prev, ...loadedPosts])
        setCurrentPage(prev => prev + 1)

        // Check if there are more posts to load
        if (endIdx >= blogPosts.length) {
          setHasMore(false)
        }
      })
      .catch(err => {
        console.error('Error loading posts:', err)
      })
  }, [blogPosts, currentPage, hasMore, loading])

  // Trigger loading when the user scrolls to the bottom
  useEffect(() => {
    if (inView && !loading && hasMore) {
      loadMorePosts()
    }
  }, [inView, loading, hasMore, loadMorePosts])

  // Load first batch of posts
  useEffect(() => {
    if (blogPosts.length > 0 && displayedPosts.length === 0) {
      loadMorePosts()
    }
  }, [blogPosts, displayedPosts.length, loadMorePosts])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'long' 
    })
  }

  const formatPeriod = (publishedDate: string, updatedDate: string) => {
    const published = formatDate(publishedDate)
    if (publishedDate !== updatedDate) {
      const updated = formatDate(updatedDate)
      return `${published} - Updated ${updated}`
    }
    return published
  }

  if (loading && displayedPosts.length === 0) {
    return (
      <div className="resume">
        <div className="container">
          <h1>Blog</h1>
          <h2>Thoughts on AI, automation, and technology</h2>
          <div className="loading">
            <p>Loading blog posts...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="resume">
        <div className="container">
          <h1>Blog</h1>
          <h2>Thoughts on AI, automation, and technology</h2>
          <div className="error">
            <p>Error loading blog posts: {error}</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="resume">
      <div className="container">
        <h1>Blog</h1>
        <h2>Thoughts on AI, automation, and technology</h2>
        
        <section className="resume-section">
          <h3>Blog Posts</h3>
          {blogPosts.length === 0 ? (
            <div className="no-experience">
              <p>No blog posts available yet.</p>
            </div>
          ) : (
            <div className="experience-items">
              {displayedPosts.map((post) => (
                <div key={post.id} className="experience-item">
                  <div className="experience-header">
                    <h4 className="experience-title">{post.title}</h4>
                    <span className="experience-company">
                      {post.authors.join(' & ')}
                    </span>
                    <span className="experience-period">
                      {formatPeriod(post.publishedDate, post.updatedDate)}
                    </span>
                  </div>
                  {post.tags && post.tags.length > 0 && (
                    <div className="experience-technologies">
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginTop: '1rem' }}>
                        {post.tags.map((tag, index) => (
                          <span 
                            key={index} 
                            style={{ 
                              fontSize: '0.75rem',
                              padding: '0.3rem 0.8rem',
                              background: 'rgba(124, 58, 237, 0.1)',
                              color: 'var(--secondary-color)',
                              border: '1px solid rgba(124, 58, 237, 0.2)',
                              borderRadius: '4px',
                              fontWeight: '500'
                            }}
                            title={tag}
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                  <div className="experience-description">
                    {post.isLoaded ? (
                      <div className="blog-post-content">
                        <ReactMarkdown
                          remarkPlugins={[remarkGfm]}
                          rehypePlugins={[rehypeRaw, rehypeSanitize]}
                        >
                          {post.content}
                        </ReactMarkdown>
                      </div>
                    ) : (
                      <p>Loading content...</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Infinite scroll trigger */}
          {hasMore && !loading && (
            <div ref={loadMoreRef} style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>
              <p>Loading more posts...</p>
            </div>
          )}

          {!hasMore && displayedPosts.length > 0 && (
            <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>
              <p>You've reached the end!</p>
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

export default Blog
