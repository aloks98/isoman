# Linting and Formatting

This project uses automated linting and formatting for both backend and frontend.

## Tools

| Component | Tool | Purpose |
|-----------|------|---------|
| Backend | golangci-lint | Comprehensive Go linter |
| Backend | gofumpt | Stricter Go formatter |
| Frontend | Biome | Fast linter + formatter for TypeScript/React |

---

## Commands

### Backend

```bash
cd backend
make fmt          # Format code
make fmt-check    # Check formatting (CI)
make lint         # Run linter
make check        # All checks: fmt-check + lint + test
```

### Frontend

```bash
cd ui
bun run format     # Format code
bun run check      # Lint and auto-fix
bun run check-only # Check only (CI)
```

---

## Pre-commit Hooks

Pre-commit hooks run automatically via Husky when you commit. They **check only** (no auto-fix):

1. **Backend (.go files)**: `gofumpt` + `golangci-lint`
2. **Frontend (.ts, .tsx, .js, .jsx, .json, .css)**: Biome CI check

### If commit fails:

```bash
# Fix backend
cd backend && make fmt
cd backend && make lint  # Fix errors manually

# Fix frontend
cd ui && bun run check

# Stage fixes and commit again
git add .
git commit -m "your message"
```

---

## Configuration

| File | Purpose |
|------|---------|
| `backend/.golangci.yml` | Go linter configuration |
| `ui/biome.json` | TypeScript/React linter + formatter |

---

## CI/CD Integration

CI pipelines run the same check-only commands:

```bash
# Backend
cd backend && make check

# Frontend
cd ui && bun run check-only
```

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "golangci-lint: command not found" | Add `$(go env GOPATH)/bin` to PATH |
| "gofumpt: command not found" | Run `cd backend && make install-tools` |
| Pre-commit hook not running | Run `bun install` in project root |

### Skip hooks (not recommended)

```bash
git commit --no-verify -m "your message"
```

---

## IDE Setup (VS Code)

### Backend (Go)

Install **Go** extension (golang.go) and add to settings:

```json
{
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "[go]": { "editor.formatOnSave": true }
}
```

### Frontend (TypeScript/React)

Install **Biome** extension (biomejs.biome) and add to settings:

```json
{
  "[typescript]": { "editor.defaultFormatter": "biomejs.biome" },
  "[typescriptreact]": { "editor.defaultFormatter": "biomejs.biome" },
  "editor.formatOnSave": true
}
```

---

## Resources

- [golangci-lint](https://golangci-lint.run/)
- [gofumpt](https://github.com/mvdan/gofumpt)
- [Biome](https://biomejs.dev/)
