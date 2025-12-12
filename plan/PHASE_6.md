# Phase 6: Docker Deployment

## Goal
Containerize the application for easy deployment.

## Tasks

### 6.1 Backend Dockerfile (`backend/Dockerfile`)

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server .

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### 6.2 Frontend Dockerfile (`frontend/Dockerfile`)

```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runtime - serve with nginx
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### 6.3 Frontend nginx.conf

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://backend:8080;
    }

    location /ws {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location /images/ {
        proxy_pass http://backend:8080;
    }
}
```

### 6.4 Docker Compose (`docker-compose.yml`)

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    environment:
      - PORT=8080
      - DATA_DIR=/data
    volumes:
      - iso-data:/data
    expose:
      - "8080"

  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  iso-data:
```

### 6.5 Alternative: Single Container

For simpler deployment, embed frontend in backend:

**Backend serves frontend:**
- Build frontend, copy dist to backend
- Gin serves static files at `/`
- Single container, single port

```go
// In routes.go
r.Static("/assets", "./frontend/dist/assets")
r.StaticFile("/", "./frontend/dist/index.html")
r.NoRoute(func(c *gin.Context) {
    c.File("./frontend/dist/index.html")
})
```

**Single Dockerfile:**
```dockerfile
# Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Build backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o server .

# Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend /app/server .
COPY --from=frontend /app/dist ./frontend/dist
EXPOSE 8080
CMD ["./server"]
```

### 6.6 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | HTTP server port |
| DATA_DIR | ./data | Base data directory |
| WORKER_COUNT | 2 | Download workers |

### 6.7 Deployment Commands

```bash
# Build and run
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop
docker-compose down

# With persistent data
docker-compose down  # keeps volumes
docker-compose down -v  # removes volumes
```

### 6.8 Volume Mounts

For accessing ISOs outside container:
```yaml
volumes:
  - ./isos:/data/isos  # Local directory mount
```

## Complete!

The application is now:
1. ✅ Downloading ISOs with progress tracking
2. ✅ Verifying checksums
3. ✅ Serving via HTTP with directory listing
4. ✅ Real-time progress via WebSocket
5. ✅ React UI for management
6. ✅ Containerized for deployment
