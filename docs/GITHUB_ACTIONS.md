# GitHub Actions CI/CD

This project uses GitHub Actions for continuous integration and deployment.

## Workflows Overview

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:** Pull requests and pushes to `master`

**Smart Path-Based Execution:**
- **Backend files** (`backend/**`, `go.mod`, `go.sum`) → Backend lint + tests
- **Frontend files** (`ui/**`, `package.json`, `bun.lock`) → Frontend lint
- **Docker files** (`Dockerfile`, `.dockerignore`) → Docker build
- **Push to master** → Always runs all jobs

**Jobs:**
1. **Detect Changes** - Analyzes changed files using path filters
2. **Backend Lint** - Check formatting (`gofumpt`) and run `golangci-lint`
3. **Backend Tests** - Run Go tests with race detector
4. **Frontend Lint** - Check TypeScript/React code with Biome
5. **Build** - Build Docker image (validate only, no push)

---

### 2. Release Workflow (`.github/workflows/release.yml`)

**Trigger:** Manual dispatch via GitHub UI

**Inputs:**
- `version` (required): Semver format (e.g., `1.2.3`)
- `release_title` (optional): Custom release title

**Actions:**
1. Validates semver format and checks tag doesn't exist
2. Updates `package.json` (root and UI) and creates `VERSION` file
3. Commits to master with `[skip ci]`
4. Creates and pushes git tag → triggers Publish workflow

---

### 3. Publish Workflow (`.github/workflows/publish.yml`)

**Trigger:** Tag push matching `v*` pattern (usually from Release workflow)

**Jobs:**
1. **Lint and Test** - Same checks as CI (backend + frontend)
2. **Build and Push** - Build Docker image and push to Docker Hub with version + `latest` tags
3. **Create GitHub Release** - Auto-generated release notes + installation instructions

---

## Usage

### Creating a Release

1. Go to **Actions → Release → Run workflow**
2. Enter version (e.g., `1.2.3`) and optional title
3. Click **Run workflow**
4. Monitor both Release and Publish workflows in Actions tab

---

## Workflow Diagrams

### CI Workflow (Pull Request)
```
Pull Request
  └─ Detect Changes
       ├─ Backend files → Backend Lint + Tests
       ├─ Frontend files → Frontend Lint
       └─ Any changes → Build Docker Image (no push)
```

### CI Workflow (Push to Master)
```
Push to Master
  ├─ Backend Lint
  ├─ Backend Tests
  └─ Frontend Lint
       └─ Build Docker Image (no push)
```

### Release + Publish Workflow
```
Manual Trigger (version: 1.2.3)
  └─ Update Version + Create Tag
       └─ [Triggers Publish Workflow]
            └─ Lint and Test
                 └─ Build and Push Docker
                      └─ Create GitHub Release
```

---

## Version Management

**Version files updated by Release workflow:**
- `package.json` (root)
- `ui/package.json`
- `VERSION` (backend)

**Manual version update:**
```bash
./scripts/update-version.sh 1.2.3
```

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Version already exists" | Use a new version number |
| Publish workflow doesn't trigger | Check tag was created and matches `v*` pattern |
| Docker push "unauthorized" | Verify `DOCKER_USERNAME` and `DOCKER_TOKEN` secrets |
| "golangci-lint not found" | Workflow runs `make install-tools` automatically |

---

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Semantic Versioning](https://semver.org/)
