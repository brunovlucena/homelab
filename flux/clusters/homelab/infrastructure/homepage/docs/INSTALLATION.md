# ⚡ Quick Installation

Get up and running in 5 minutes!

## 1. Add to Shell Config

Add this line to your `~/.zshrc` or `~/.bashrc`:

```bash
source ~/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/scripts/git-helpers.sh
```

## 2. Reload Shell

```bash
source ~/.zshrc  # or source ~/.bashrc
```

## 3. Verify

```bash
ghelp
```

You should see:
```
🚀 Git Helper Functions for Homepage

Branch Creation:
  gfeature <desc>     - Create feature branch from develop
  gbugfix <desc>      - Create bugfix branch from develop
  ghotfix <desc>      - Create hotfix branch from main
  ...
```

## Done! 🎉

Now you can use commands like:
- `gfeature add search` - Create feature branch
- `gcommit feat add search` - Commit with convention
- `gbump 1.1.0` - Bump version
- `ginfo` - Check status

See [Git Helpers Guide](./GIT_HELPERS_GUIDE.md) for full documentation.

