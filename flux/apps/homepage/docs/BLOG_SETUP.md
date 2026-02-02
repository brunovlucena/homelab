# Blog Setup Instructions

## Installation

To complete the blog setup, you need to install the required markdown rendering packages:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/homepage/src/frontend
npm install react-markdown remark-gfm rehype-raw rehype-sanitize
```

Or if using yarn:

```bash
yarn add react-markdown remark-gfm rehype-raw rehype-sanitize
```

Or if using pnpm:

```bash
pnpm add react-markdown remark-gfm rehype-raw rehype-sanitize
```

## What was created

1. **Blog Page** (`src/pages/Blog.tsx`) - Lists all blog posts
2. **BlogPost Page** (`src/pages/BlogPost.tsx`) - Displays individual blog posts
3. **Blog Styles** (`src/Blog.css`) - Styling for blog pages
4. **Routes** - Added `/blog` and `/blog/:slug` routes to App.tsx
5. **Navigation** - Added Blog link to Header component

## Blog Post Content

Blog posts are stored as markdown files in `public/blog-posts/` and managed via `metadata.json`.

## Features

- **File-based blog posts** - Posts stored as `.md` files with frontmatter
- **Draft/Published control** - Control post visibility via `published` field
- **Infinite scroll** - Lazy loads posts as user scrolls (5 posts per page)
- **Lazy loading** - Post content loaded on-demand for performance
- **Smart caching** - Metadata cached in localStorage (5-minute TTL)
- **Security** - HTML sanitization via rehype-sanitize prevents XSS attacks
- **Responsive design**
- **Markdown rendering** with GitHub Flavored Markdown support
- **Syntax highlighting** for code blocks
- **Tag system**
- **Author attribution**
- **Reading time estimation**
- **Semantic HTML** with proper time elements and accessibility
- **Clean, modern UI** matching the existing site theme

## Performance & Security Features

### Infinite Scroll
The blog automatically loads 5 posts at a time as you scroll. This prevents loading all posts on initial page load, improving performance for blogs with many posts.

### Lazy Loading
Post content is fetched on-demand only when displayed. Metadata is loaded immediately, but the actual markdown content is fetched progressively as users scroll.

### Smart Caching
Blog metadata is cached in `localStorage` with a 5-minute TTL. This reduces server requests and improves perceived performance on repeat visits.

### Security
- **HTML Sanitization**: All markdown content is sanitized via `rehype-sanitize` to prevent XSS attacks
- **Safe rendering**: Even if malicious HTML is in markdown files, it's stripped before rendering

### Constants
Hard-coded paths have been extracted to constants at the top of `Blog.tsx`:
- `BLOG_POSTS_BASE_URL`: Base URL for blog posts
- `METADATA_FILE`: Metadata JSON filename
- `POSTS_PER_PAGE`: Number of posts to load per page
- `CACHE_DURATION_MS`: Cache time-to-live

## Next Steps

1. Run `npm install` to install the required packages (including `rehype-sanitize`)
2. Start the development server with `npm run dev`
3. Navigate to `/blog` to see your blog with infinite scroll
4. The "Understanding vs. Knowledge" post is available, now enhanced with observability examples

## Adding New Posts

### 1. Create the Markdown File

Create a new `.md` file in `public/blog-posts/`:

```bash
touch public/blog-posts/my-new-post.md
```

Add frontmatter and content:

```markdown
---
title: "My Post Title"
slug: "my-post-slug"
authors:
  - Bruno Lucena
tags:
  - Tag1
  - Tag2
publishedDate: 2025-12-17
readTime: "10 min read"
published: false
---

Your content here...
```

### 2. Update Metadata

Add an entry to `public/blog-posts/metadata.json`:

```json
{
  "id": "3",
  "slug": "my-post-slug",
  "title": "My Post Title",
  "filename": "my-new-post.md",
  "authors": ["Bruno Lucena"],
  "tags": ["Tag1", "Tag2"],
  "publishedDate": "2025-12-17T12:00:00Z",
  "updatedDate": "2025-12-17T12:00:00Z",
  "readTime": "10 min read",
  "published": false
}
```

### 3. Control Visibility

**Draft Mode** (post exists but not shown on blog):
```json
"published": false
```

**Published Mode** (post visible on blog):
```json
"published": true
```

Or omit the field entirely (defaults to `true`).

### 4. Optional: Auto-generate Metadata

You can use the script to generate metadata from markdown frontmatter:

```bash
node scripts/generate-blog-metadata.mjs
```

Note: This will scan all `.md` files in `public/blog-posts/` and generate `metadata.json`.

## Post Status Control

The blog system supports draft/published workflow:

- **Draft posts** (`published: false`): Stored in repo but not shown on blog
- **Published posts** (`published: true`): Visible on blog
- **Default behavior**: If `published` field is omitted, post is visible (defaults to `true`)

This allows you to:
- Commit draft posts to version control
- Review and edit before publishing
- Publish by simply changing `published: false` → `published: true`
