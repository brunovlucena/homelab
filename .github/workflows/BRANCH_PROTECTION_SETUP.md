# 🔒 Branch Protection Setup Guide

This guide explains how to configure GitHub branch protection rules to require all tests to pass before merging pull requests to `main`.

## 📋 Overview

The repository now has comprehensive test workflows that run automatically on every pull request:

- **🏠 Homepage Tests**: Unit tests for Go API and React frontend
- **🤖 Agent SRE Tests**: Python unit tests with pytest
- **🤖 Agent Jamie Tests**: Python unit tests with pytest
- **🔒 PR Test Gate**: Unified workflow that ensures all relevant tests pass

## 🎯 Required Branch Protection Rules

To enforce that tests must pass before merging, follow these steps:

### Step 1: Navigate to Branch Protection Settings

1. Go to your repository on GitHub: `https://github.com/brunovlucena/homelab`
2. Click on **Settings** tab
3. In the left sidebar, click on **Branches** (under "Code and automation")
4. Click **Add branch protection rule** or edit the existing rule for `main`

### Step 2: Configure the Branch Protection Rule

Set the following options:

#### Branch Name Pattern
```
main
```

#### Protect Matching Branches

Check these options:

- ✅ **Require a pull request before merging**
  - ✅ Require approvals: `1` (or more based on your preference)
  - ✅ Dismiss stale pull request approvals when new commits are pushed
  - ✅ Require review from Code Owners (optional)

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - **Add the following required status checks:**
    - `✅ All Tests Passed` (from PR Test Gate workflow)
    - `Backend Unit Tests` (from Homepage Tests)
    - `Frontend Unit Tests` (from Homepage Tests)
    - `Build Verification` (from Homepage Tests)
    - `Agent SRE Tests` (from Agent SRE Tests)
    - `Agent Jamie Tests` (from Agent Jamie Tests)

- ✅ **Require conversation resolution before merging** (recommended)

- ✅ **Require signed commits** (optional, but recommended for security)

- ✅ **Require linear history** (optional, keeps history clean)

- ✅ **Include administrators** (ensures rules apply to everyone)

- ✅ **Restrict who can push to matching branches** (optional)

- ❌ **Allow force pushes** - Keep this DISABLED
- ❌ **Allow deletions** - Keep this DISABLED

### Step 3: Save the Rule

Click **Create** or **Save changes** at the bottom of the page.

## 🔍 How It Works

Once branch protection is enabled:

1. **Developer creates a PR** → Workflows automatically trigger
2. **Tests run in parallel**:
   - If Homepage changed → Homepage tests run
   - If Agent SRE changed → Agent SRE tests run
   - If Agent Jamie changed → Agent Jamie tests run
3. **PR Test Gate validates** all relevant tests passed
4. **GitHub blocks merge** if any test fails
5. **Developer fixes issues** → Tests rerun automatically
6. **All tests pass** → PR can be merged ✅

## 📊 Status Checks Explained

### Required Status Checks

These must ALL pass before merging:

| Status Check | Workflow | Purpose |
|-------------|----------|---------|
| `✅ All Tests Passed` | PR Test Gate | Main gate that validates all relevant tests |
| `Backend Unit Tests` | Homepage Tests | Go unit tests with coverage |
| `Frontend Unit Tests` | Homepage Tests | React/TypeScript tests |
| `Build Verification` | Homepage Tests | Ensures code compiles |
| `Metrics Tests` | Homepage Tests | Metrics package tests |
| `Agent SRE Tests` | Agent SRE Tests | Python tests for SRE agent |
| `Agent Jamie Tests` | Agent Jamie Tests | Python tests for Jamie agent |

### How Tests Are Triggered

Tests only run when relevant files change:

```yaml
# Homepage tests run when:
- flux/clusters/homelab/infrastructure/homepage/** changes

# Agent SRE tests run when:
- flux/clusters/homelab/infrastructure/agent-sre/** changes

# Agent Jamie tests run when:
- flux/clusters/homelab/infrastructure/agent-jamie/** changes
- flux/clusters/homelab/infrastructure/jamie/** changes
```

## 🚫 What Gets Blocked

With branch protection enabled, these actions will be **blocked**:

1. ❌ Merging a PR with failing tests
2. ❌ Merging a PR without required approvals
3. ❌ Direct pushes to `main` branch (if configured)
4. ❌ Force pushes to `main` branch
5. ❌ Deleting the `main` branch

## ✅ Best Practices

### For Contributors

1. **Always create a branch** for your changes
   ```bash
   git checkout -b fix/issue-description
   ```

2. **Run tests locally** before pushing
   ```bash
   # Homepage
   cd flux/clusters/homelab/infrastructure/homepage
   make test
   
   # Agent SRE
   cd flux/clusters/homelab/infrastructure/agent-sre
   uv run pytest tests/
   
   # Agent Jamie
   cd flux/clusters/homelab/infrastructure/agent-jamie
   uv run pytest tests/
   ```

3. **Keep PRs focused** - one feature/fix per PR

4. **Watch the checks** - GitHub will show status in the PR

5. **Fix failures quickly** - tests run automatically on each push

### For Reviewers

1. **Wait for green checks** before reviewing
2. **Review test coverage** in the PR summary
3. **Verify test changes** if test files were modified
4. **Request changes** if tests are insufficient

## 🔄 Updating Branch Protection

To modify protection rules later:

1. Go to **Settings** → **Branches**
2. Click the **Edit** button next to the `main` rule
3. Make your changes
4. Click **Save changes**

## 🆘 Troubleshooting

### "Required status checks are not showing up"

**Cause**: Status checks only appear after they run at least once.

**Solution**: 
1. Create a test PR that modifies relevant files
2. Let workflows run
3. Go back to branch protection settings
4. The status checks will now appear in the list

### "I need to merge an urgent fix but tests are failing"

**Options**:
1. **Best**: Fix the tests (recommended)
2. **If urgent**: Temporarily disable branch protection
   - Go to Settings → Branches → Edit rule
   - Uncheck "Require status checks to pass"
   - Merge your fix
   - **Immediately re-enable protection**

### "Tests pass locally but fail in CI"

**Common causes**:
- Different Python/Go/Node versions
- Missing environment variables
- Platform-specific issues (Mac vs Linux)

**Solution**:
- Check workflow logs for specific errors
- Ensure local environment matches CI versions
- Add missing dependencies to requirements/package files

## 📚 Additional Resources

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [Status Checks Documentation](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks)
- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)

## 📝 Notes

- Branch protection rules apply to **all users** including administrators (if "Include administrators" is checked)
- Status checks run in parallel for faster feedback
- Failed checks will show detailed logs in the Actions tab
- You can bypass checks temporarily with admin rights (not recommended)

---

**Last Updated**: October 11, 2025  
**Maintained By**: @brunovlucena


