# Dynamic Dates Setup for Blog Posts

## 🎉 What Changed

Your blog now has **DYNAMIC FUCKING DATES**! No more hardcoded bullshit.

## How It Works

### 1. **Frontmatter in Markdown Files**
Each blog post now has YAML frontmatter at the top:

```markdown
---
title: "Your Post Title"
slug: "your-post-slug"
authors:
  - Bruno Lucena
tags:
  - AI
  - Philosophy
publishedDate: 2025-12-17
readTime: "8 min read"
---

Your content starts here...
```

### 2. **Auto-Generated metadata.json**
The `metadata.json` file is now **automatically generated** by a script:

- **Location:** `scripts/generate-blog-metadata.mjs`
- **What it does:**
  - Reads all `.md` files in `public/blog-posts/`
  - Parses frontmatter for metadata
  - Uses `publishedDate` from frontmatter
  - **DYNAMICALLY sets `updatedDate` from file modification time**
  - Generates `metadata.json` automatically

### 3. **Integrated into Build Process**
The script runs automatically:

```json
{
  "scripts": {
    "build": "node scripts/generate-blog-metadata.mjs && tsc && vite build",
    "generate-blog-metadata": "node scripts/generate-blog-metadata.mjs"
  }
}
```

## Usage

### Adding a New Blog Post

1. **Create a markdown file** with frontmatter:
```bash
touch public/blog-posts/my-new-post.md
```

2. **Add frontmatter** at the top:
```markdown
---
title: "My New Post"
slug: "my-new-post"
authors:
  - Bruno Lucena
tags:
  - Whatever
publishedDate: 2025-12-17
readTime: "5 min read"
---

Content here...
```

3. **Generate metadata:**
```bash
npm run generate-blog-metadata
```

4. **Done!** The `updatedDate` is automatically set to the file's last modification time.

### Updating an Existing Post

1. **Edit the markdown file** - just save your changes
2. **Regenerate metadata:**
```bash
npm run generate-blog-metadata
```
3. **The `updatedDate` updates automatically** based on file modification time!

## What's Different

### Before (HARDCODED 💩)
```json
{
  "publishedDate": "2024-12-17T19:22:00Z",
  "updatedDate": "2024-12-17T19:40:00Z"  // MANUALLY SET, ALWAYS WRONG
}
```

### After (DYNAMIC 🚀)
```markdown
---
publishedDate: 2025-12-17  // Set once in frontmatter
---
```

```javascript
// updatedDate is AUTOMATICALLY generated from file modification time
const stats = fs.statSync(filePath);
const updatedDate = stats.mtime.toISOString();  // ALWAYS CURRENT!
```

## Files Modified

1. **`public/blog-posts/understanding-vs-knowledge.md`** - Added frontmatter
2. **`scripts/generate-blog-metadata.mjs`** - NEW! Auto-generation script
3. **`package.json`** - Added `gray-matter` dependency and scripts
4. **`public/blog-posts/README.md`** - Updated documentation

## Requirements

- `gray-matter@^4.0.3` - For parsing frontmatter (added to devDependencies)
- Node.js - For running the script

## Commands

```bash
# Generate metadata manually
npm run generate-blog-metadata

# Build (automatically generates metadata first)
npm run build

# Development
npm run dev
```

## Notes

- **DO NOT manually edit `metadata.json`** - it's auto-generated
- The `updatedDate` is always current based on file system
- The `publishedDate` is set once in frontmatter and never changes
- Every time you save a markdown file, the next metadata generation will update the `updatedDate`

---

**No more hardcoded dates! 🎉**
