# Linting and Formatting

This project uses automated linting and formatting tools for both backend (Go) and frontend (TypeScript/React).

## Tools

### Backend (Go)

- **golangci-lint** - Comprehensive linter that runs multiple Go linters
- **gofumpt** - Stricter version of `gofmt` for consistent formatting

### Frontend (TypeScript/React)

- **Biome** - Fast linter and formatter (alternative to ESLint + Prettier)

## Setup

### Install Tools

#### Backend

```bash
cd backend
make install-tools
```

This installs:
- `golangci-lint` - Code quality checker
- `gofumpt` - Code formatter

### Pre-commit Hooks

Pre-commit hooks are automatically set up via [Husky](https://typicode.github.io/husky/) when you run `bun install` in the project root.

The hooks will:
1. **Format** backend code with `gofumpt`
2. **Check** frontend code with Biome

## Usage

### Backend

#### Format Code

```bash
cd backend
make fmt
```

#### Check Formatting

```bash
cd backend
make fmt-check
```

#### Run Linter

```bash
cd backend
make lint
```

#### Run All Checks

```bash
cd backend
make check  # Runs lint + fmt-check + tests
```

### Frontend

#### Format Code

```bash
cd ui
bun run format
```

#### Lint and Fix

```bash
cd ui
bun run check
```

## Configuration

### Backend

Configuration is in `backend/.golangci.yml`:
- Enabled linters: errcheck, gosimple, govet, staticcheck, unused, gofumpt, and more
- Cyclomatic complexity threshold: 20
- Excludes test files from some checks

### Frontend

Configuration is in `ui/biome.json`:
- TypeScript/React specific rules
- Auto-formatting on save (if IDE is configured)

## Pre-commit Hook Behavior

When you run `git commit`, the hook will **CHECK ONLY** (not auto-fix):

1. **Backend (.go files)**:
   - Run `gofumpt` to check formatting
   - **Fails if code is not formatted**

2. **Frontend (.ts, .tsx, .js, .jsx, .json, .css files)**:
   - Run Biome check (lint + format check)
   - **Fails if code has linting or formatting issues**

### If the commit fails:

1. **Format your code manually**:
   ```bash
   # Backend
   cd backend && make fmt

   # Frontend
   cd ui && bun run check
   ```

2. **Stage the formatted files**:
   ```bash
   git add .
   ```

3. **Commit again**:
   ```bash
   git commit -m "your message"
   ```

This ensures:
- Developers are aware of formatting issues before committing
- CI validation is consistent with local pre-commit checks
- No surprise auto-formatting in commits

## CI/CD Integration

For CI/CD pipelines, use the same check-only commands as pre-commit:

### Backend
```bash
cd backend
make fmt-check  # Check formatting (fails if not formatted)
make lint       # Run linters (fails if issues found)
make test       # Run tests
# Or run all checks together:
make check      # fmt-check + lint + test
```

### Frontend
```bash
cd ui
bun run check-only  # Check formatting and linting (fails if issues found)
# Or auto-fix issues (for local development):
bun run check       # Auto-fix and apply changes
```

### Complete CI Pipeline

```bash
# Install backend tools
cd backend && make install-tools

# Check backend
cd backend && make check  # fmt-check + lint + test

# Check frontend
cd ui && bun run check-only
```

## Troubleshooting

### "golangci-lint: command not found"

Ensure `$(go env GOPATH)/bin` is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add this to your `.bashrc`, `.zshrc`, or equivalent shell config file.

### "gofumpt: command not found"

Run:
```bash
cd backend
make install-tools
```

### Pre-commit hook not running

1. Ensure Husky is installed:
   ```bash
   cd /path/to/isoman
   bun install
   ```

2. Check if `.husky/pre-commit` exists and is executable:
   ```bash
   ls -la .husky/pre-commit
   chmod +x .husky/pre-commit
   ```

### Skip Pre-commit Hooks (Not Recommended)

In rare cases, you can skip hooks with:
```bash
git commit --no-verify -m "your message"
```

**Note**: This bypasses quality checks and should only be used in emergencies.

## IDE Integration

### VS Code

#### Backend (Go)

Install extensions:
- **Go** (golang.go)

Add to `.vscode/settings.json`:
```json
{
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

#### Frontend (TypeScript/React)

Install extension:
- **Biome** (biomejs.biome)

Add to `.vscode/settings.json`:
```json
{
  "[javascript]": {
    "editor.defaultFormatter": "biomejs.biome"
  },
  "[typescript]": {
    "editor.defaultFormatter": "biomejs.biome"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "biomejs.biome"
  },
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "quickfix.biome": "explicit",
    "source.organizeImports.biome": "explicit"
  }
}
```

### Other IDEs

- **GoLand/IntelliJ IDEA**: Enable `gofumpt` in Settings → Go → Gofmt
- **Neovim/Vim**: Use `null-ls` or `ALE` with golangci-lint and gofumpt

## Best Practices

1. **Run formatting before committing**:
   ```bash
   cd backend && make fmt
   cd ui && bun run format
   ```

2. **Fix linting issues incrementally**: Don't disable all linters. Fix issues as you work on code.

3. **Keep tools updated**:
   ```bash
   # Backend tools
   cd backend && make install-tools

   # Frontend tools
   cd ui && bun update
   ```

4. **Review linting configuration** as the project evolves. Adjust `.golangci.yml` and `biome.json` based on team preferences.

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [gofumpt GitHub](https://github.com/mvdan/gofumpt)
- [Biome Documentation](https://biomejs.dev/)
- [Husky Documentation](https://typicode.github.io/husky/)
- [lint-staged Documentation](https://github.com/okonet/lint-staged)
