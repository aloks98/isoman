# GitHub Actions CI/CD

This project uses GitHub Actions for continuous integration and deployment.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Pull requests to `master`
- Pushes to `master`

**Smart Path-Based Execution:**

The workflow intelligently runs jobs based on which files changed:

**On Pull Requests:**
- **Backend files changed** (`backend/**`, `go.mod`, `go.sum`) → Run backend lint + tests only
- **Frontend files changed** (`ui/**`, `package.json`, `bun.lock`) → Run frontend lint only
- **Docker files changed** (`Dockerfile`, `.dockerignore`) → Run Docker build only
- **Multiple areas changed** → Run all relevant jobs
- **All jobs must pass** before merge

**On Push to Master:**
- Always runs **all jobs** (backend, frontend, and Docker build)

**Jobs:**

1. **Detect Changes**
   - Analyzes changed files using path filters
   - Sets outputs for backend/frontend/docker changes

2. **Backend Lint** (conditional)
   - Runs if backend files changed OR on push to master
   - Check Go code formatting with `gofumpt`
   - Run `golangci-lint`

3. **Backend Tests** (conditional)
   - Runs if backend files changed OR on push to master
   - Run all Go tests with race detector

4. **Frontend Lint** (conditional)
   - Runs if frontend files changed OR on push to master
   - Check TypeScript/React code with Biome CI
   - Verify formatting and linting

5. **Build** (conditional)
   - Runs if any source files changed OR on push to master
   - Only runs if all previous jobs succeeded (or were skipped)
   - Build Docker image to verify successful build (does NOT push)

**Features:**
- ⚡ **Fast PR checks**: Only runs relevant jobs
- 🔄 **Dependency caching**:
  - Go modules (`~/go/pkg/mod`) and build artifacts (`~/.cache/go-build`)
  - Bun dependencies (`~/.bun/install/cache` and `ui/node_modules`)
- 🚀 Parallel execution of independent jobs
- 🐳 Docker layer caching for faster builds
- 🔍 Always runs full suite on master branch
- ✅ **Build validation only**: Docker image is built but not pushed

---

### 2. Release Workflow (`.github/workflows/release.yml`)

**Trigger:**
- Manual dispatch via GitHub UI
- Requires version input (e.g., `1.2.3`)
- Optional release title

**Inputs:**
- `version` (required): Release version in semver format (e.g., `1.2.3`)
- `release_title` (optional): Custom release title (defaults to "Release vX.Y.Z")

**Job:**

1. **Update Version and Create Tag**
   - Validates semver format (X.Y.Z)
   - Checks if tag already exists (fails if it does)
   - Updates `package.json` (root and UI)
   - Creates `VERSION` file in project root
   - Commits changes to `master` with `[skip ci]`
   - **Creates and pushes git tag** (e.g., `v1.2.3`) with optional custom title
   - **Triggers Publish workflow** automatically

**Key Features:**
- ⚡ **Fast execution**: Only updates version files and creates tag
- 🔒 **Version validation**: Ensures semver format and unique tags
- 🚫 **Skip CI on version commit**: Uses `[skip ci]` to avoid triggering CI workflow
- 🎯 **Automatic publish**: Tag creation triggers the Publish workflow
- 📝 **Custom title**: Optionally provide custom title for GitHub release

---

### 3. Publish Workflow (`.github/workflows/publish.yml`)

**Trigger:**
- Automatic trigger on tag push matching `v*` pattern
- Typically triggered by Release workflow

**Jobs:**

1. **Lint and Test**
   - Runs same checks as CI workflow (backend + frontend in parallel)
   - Must pass before proceeding to build

2. **Build and Push**
   - Depends on lint and test jobs
   - Extracts version from tag
   - Builds Docker image with version
   - Pushes to Docker Hub with two tags:
     - `<username>/isoman:<version>` (e.g., `1.2.3`)
     - `<username>/isoman:latest`
   - Extracts custom title from tag (if provided)
   - Creates GitHub Release with auto-generated commit history and installation instructions

**Key Features:**
- 🔒 **Protected by tests**: Docker push only happens if all tests pass
- 🎯 **Build from tag**: Ensures Docker image is built from tagged commit
- 📦 **Dual tagging**: Pushes both version-specific and `latest` tags
- 🔄 **Dependency caching**: Go and Bun dependencies cached for faster builds
- 📝 **Auto-generated release notes**: GitHub automatically generates release notes from commits since last release
- 🐳 **Complete installation guide**: Includes Docker and Docker Compose installation instructions

---

## Setup Instructions

### 1. Docker Hub Setup

Create a Docker Hub account and access token:

1. **Create Docker Hub Account**: https://hub.docker.com/signup
2. **Create Access Token**:
   - Go to Account Settings → Security → New Access Token
   - Name: `GitHub Actions`
   - Permissions: Read, Write, Delete
   - **Copy the token** (you won't see it again!)

### 2. GitHub Secrets Configuration

Add the following secrets to your GitHub repository:

**Settings → Secrets and variables → Actions → New repository secret**

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `DOCKER_USERNAME` | Your Docker Hub username | `johndoe` |
| `DOCKER_TOKEN` | Docker Hub access token | `dckr_pat_abc123...` |

### 3. Enable GitHub Actions

1. Go to your repository → **Actions** tab
2. If prompted, click **"I understand my workflows, go ahead and enable them"**
3. Workflows will now run automatically on push/PR

---

## Usage

### Running CI (Automatic)

CI runs automatically on:
- **Pull Requests**: Validates changes before merging
- **Push to Master**: Ensures master branch is always healthy

**No action needed** - just push code or create a PR!

### Creating a Release (Manual)

1. **Navigate to Actions**:
   ```
   Repository → Actions → Release → Run workflow
   ```

2. **Fill in the form**:
   - **Branch**: `master` (default)
   - **Release version** (required): Enter version (e.g., `1.2.3`)
   - **Release title** (optional): Custom title for GitHub release

   **Examples:**
   - **Default title**: Leave empty, will use "Release v1.2.3"
   - **Custom title**: `ISOman v1.2.3 - Performance Improvements`

3. **Click "Run workflow"**

4. **What happens**:
   - Release workflow updates version files and creates tag
   - Tag creation automatically triggers Publish workflow
   - Publish workflow runs lint, tests, builds, and pushes Docker image
   - GitHub release created with:
     - Custom title (if provided) or default title
     - Auto-generated commit history since last release
     - Docker and Docker Compose installation instructions

5. **Monitor progress**:
   - Watch both workflows in the Actions tab
   - Release workflow completes quickly
   - Publish workflow runs lint, tests, and build

6. **Verify release**:
   - Check Releases page for new GitHub release
   - Release includes auto-generated commits and installation guide
   - Check Docker Hub for new image tags:
     ```bash
     docker pull <username>/isoman:1.2.3
     docker pull <username>/isoman:latest
     ```

---

## Workflow Diagrams

### CI Workflow (Pull Request - Smart Execution)
```
Pull Request (detects which files changed)
  └─ Detect Changes (path filter)
       ├─ Backend files → Backend Lint + Backend Tests (parallel)
       ├─ Frontend files → Frontend Lint
       ├─ Docker files → (waits for other jobs)
       └─ Any changes → Build Docker Image (validate only, no push)
```

### CI Workflow (Push to Master - Full Suite)
```
Push to Master
  ├─ Backend Lint (parallel)
  ├─ Backend Tests (parallel)
  └─ Frontend Lint (parallel)
       └─ Build Docker Image (validate only, no push)
```

### Release Workflow
```
Manual Trigger (version: 1.2.3, optional: title)
  └─ Update Version and Create Tag
       ├─ Validate version format
       ├─ Check tag doesn't exist
       ├─ Update package.json files + VERSION
       ├─ Commit to master [skip ci]
       └─ Create and push git tag (v1.2.3) with optional custom title
            └─ [Automatically triggers Publish Workflow]
```

### Publish Workflow
```
Tag Push (v1.2.3)
  └─ Lint and Test (backend + frontend in parallel)
       └─ Build and Push Docker Image
            ├─ Build from tag
            ├─ Push: 1.2.3
            ├─ Push: latest
            └─ Create GitHub Release
                 ├─ Title: Custom (if provided) or default "Release v1.2.3"
                 ├─ Body: Auto-generated commit history + Installation instructions
                 └─ Includes Docker & Docker Compose examples
```

---

## Version Management

### Version File Locations

The release workflow updates version in:
- `package.json` (root)
- `ui/package.json`
- `VERSION` (backend)

### Manual Version Update (Local)

Use the version update script:

```bash
./scripts/update-version.sh 1.2.3
```

This updates all version files locally (doesn't create tags).

### Version Format

Must follow semantic versioning (semver):
- Format: `MAJOR.MINOR.PATCH`
- Examples: `1.0.0`, `1.2.3`, `2.0.0`
- Invalid: `v1.0`, `1.0`, `1.0.0-beta`

---

## Troubleshooting

### Release Workflow Fails: "Version already exists"

**Problem**: Tag already exists for this version

**Solution**:
1. Check existing tags: `git tag`
2. Use a new version number
3. Or delete the old tag (not recommended):
   ```bash
   git tag -d v1.2.3
   git push origin :refs/tags/v1.2.3
   ```

### Publish Workflow Doesn't Trigger

**Problem**: Release workflow completes but Publish workflow doesn't start

**Solution**:
1. Ensure the tag was created: `git tag`
2. Check that tag matches `v*` pattern (e.g., `v1.2.3`)
3. Verify GitHub Actions is enabled for tag events in repository settings

### Docker Push Fails: "unauthorized"

**Problem**: Docker Hub credentials are incorrect

**Solution**:
1. Verify secrets are set correctly in GitHub
2. Regenerate Docker Hub access token
3. Update `DOCKER_TOKEN` secret

### CI Fails: "golangci-lint: command not found"

**Problem**: Tools not installed properly

**Solution**: The workflow runs `make install-tools` which should install golangci-lint. Check if the Makefile exists and is correct.

### Build Fails: "manifest unknown"

**Problem**: Base image not found or Docker Hub repository doesn't exist

**Solution**:
1. Verify Dockerfile base images
2. Ensure Docker Hub repository is created (auto-created on first push)

---

## CI/CD Best Practices

### 1. Branch Protection Rules

Enable branch protection for `master`:

**Settings → Branches → Branch protection rules → Add rule**

- Branch name pattern: `master`
- ✅ Require a pull request before merging
- ✅ Require status checks to pass before merging
  - Select: `Backend Lint`, `Backend Tests`, `Frontend Lint`, `Build`
- ✅ Require branches to be up to date before merging

### 2. Pull Request Workflow

Recommended workflow:
1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and commit
3. Push and create pull request
4. **Wait for CI to pass** (all green checkmarks)
5. Request review (if needed)
6. Merge to master

### 3. Release Workflow

Recommended release process:
1. **Ensure master is healthy**: Check latest CI run is green
2. **Plan version bump**:
   - Patch (1.0.X): Bug fixes
   - Minor (1.X.0): New features (backwards compatible)
   - Major (X.0.0): Breaking changes
3. **Run release workflow** with chosen version
   - Release workflow updates version and creates tag
   - Publish workflow automatically triggered by tag
4. **Monitor both workflows**:
   - Release workflow should complete quickly
   - Publish workflow runs lint, tests, and build
5. **Test deployed image**:
   ```bash
   docker pull <username>/isoman:latest
   docker run -p 8080:8080 <username>/isoman:latest
   ```
6. **Announce release** (if public project)

### 4. Rollback Strategy

If a release has issues:

1. **Quick fix (patch release)**:
   - Fix bug on master
   - Create new patch release (e.g., 1.2.4)

2. **Full rollback**:
   ```bash
   # Pull previous working version
   docker pull <username>/isoman:1.2.2

   # Retag as latest
   docker tag <username>/isoman:1.2.2 <username>/isoman:latest
   docker push <username>/isoman:latest
   ```

---

## Monitoring

### GitHub Actions Dashboard

View workflow runs:
```
Repository → Actions
```

- ✅ Green checkmark: Success
- ❌ Red X: Failure
- 🟡 Yellow dot: In progress

### Email Notifications

GitHub sends email notifications for:
- Failed workflow runs (if you're the author)
- Workflows you triggered

Configure in: **Settings → Notifications → Actions**

### Status Badges

Add build status badge to README:

```markdown
![CI](https://github.com/<username>/isoman/workflows/CI/badge.svg)
```

---

## Cost Considerations

GitHub Actions is **free** for public repositories.

For **private repositories**:
- Free tier: 2,000 minutes/month
- Each workflow run uses minutes

**Optimize costs**:
- Use caching (already implemented)
- Cancel duplicate runs (Settings → Actions → General)
- Run only on necessary events

---

## Advanced Configuration

### Customizing Docker Image Name

Edit `.github/workflows/release.yml`:

```yaml
images: your-dockerhub-username/your-image-name
```

### Adding More Test Steps

Edit `.github/workflows/ci.yml`, add steps to `backend-test` job:

```yaml
- name: Integration tests
  run: go test -tags=integration ./...
```

### Triggering Release on Tag Push

Alternative to manual workflow - auto-release on tag:

```yaml
on:
  push:
    tags:
      - 'v*'
```

---

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Semantic Versioning](https://semver.org/)
- [Docker Hub](https://hub.docker.com/)
