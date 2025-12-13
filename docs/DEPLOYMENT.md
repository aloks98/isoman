# ISOMan Deployment Guide

Complete guide for deploying ISOMan in production.

## Deployment Options

1. **Docker (Recommended)** - Single container with embedded UI
2. **Manual** - Separate backend binary and static UI files
3. **Development** - Local development with hot reload

---

## Docker Deployment

### Prerequisites

- Docker installed (version 20.10+)
- At least 2GB RAM
- Sufficient disk space for ISOs (depends on your usage)

### Quick Start

1. **Build the image**
   ```bash
   docker build -t isoman .
   ```

2. **Run the container**
   ```bash
   docker run -d \
     --name isoman \
     -p 8080:8080 \
     -v isoman-data:/data \
     --restart unless-stopped \
     isoman
   ```

3. **Access the application**
   Open `http://localhost:8080` or `http://your-server-ip:8080`

### Configuration Options

#### Port Mapping

Change the external port (left side):
```bash
docker run -d -p 3000:8080 -v isoman-data:/data isoman
# Access at http://localhost:3000
```

#### Data Persistence

Use a named volume (recommended):
```bash
docker run -d -p 8080:8080 -v isoman-data:/data isoman
```

Use a host directory:
```bash
docker run -d -p 8080:8080 -v /path/on/host:/data isoman
```

#### Environment Variables

```bash
docker run -d \
  -p 8080:8080 \
  -v isoman-data:/data \
  -e PORT=8080 \
  -e DATA_DIR=/data \
  -e WORKER_COUNT=3 \
  isoman
```

**Available Environment Variables:**

See `backend/ENV.md` for complete list of 26 environment variables. Common ones include:
- `PORT` - HTTP server port (default: 8080)
- `DATA_DIR` - Base data directory (default: /data)
- `WORKER_COUNT` - Number of concurrent downloads (default: 2)
- `LOG_LEVEL` - Logging level (default: info)
- `DB_MAX_OPEN_CONNS` - Database connection pool size (default: 10)

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  isoman:
    image: isoman
    build: .
    container_name: isoman
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - isoman-data:/data
    environment:
      - PORT=8080
      - DATA_DIR=/data
      - WORKER_COUNT=2

volumes:
  isoman-data:
    driver: local
```

Run with:
```bash
docker-compose up -d
```

### Reverse Proxy Setup

#### Nginx

Create `/etc/nginx/sites-available/isoman`:

```nginx
server {
    listen 80;
    server_name iso.example.com;

    # Optional: Redirect to HTTPS
    # return 301 https://$server_name$request_uri;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # Large file uploads (if needed)
    client_max_body_size 0;
}
```

Enable and reload:
```bash
sudo ln -s /etc/nginx/sites-available/isoman /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### Caddy

Create `Caddyfile`:

```caddy
iso.example.com {
    reverse_proxy localhost:8080
}
```

Run:
```bash
caddy run
```

---

## Manual Deployment

### Build from Source

1. **Build the backend**
   ```bash
   cd backend
   CGO_ENABLED=0 go build -ldflags="-w -s" -o server .
   ```

2. **Build the UI**
   ```bash
   cd ui
   bun install
   bun run build
   ```

3. **Directory structure**
   ```
   /opt/isoman/
   ├── server           # Backend binary
   ├── ui/
   │   └── dist/       # Built UI files
   └── data/           # Data directory (created automatically)
   ```

### Systemd Service

Create `/etc/systemd/system/isoman.service`:

```ini
[Unit]
Description=ISOMan - Linux ISO Download Manager
After=network.target

[Service]
Type=simple
User=isoman
Group=isoman
WorkingDirectory=/opt/isoman
ExecStart=/opt/isoman/server
Restart=always
RestartSec=10

# Environment
Environment="PORT=8080"
Environment="DATA_DIR=/opt/isoman/data"
Environment="WORKER_COUNT=2"

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/isoman/data

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo useradd -r -s /bin/false isoman
sudo mkdir -p /opt/isoman/data
sudo chown -R isoman:isoman /opt/isoman
sudo systemctl daemon-reload
sudo systemctl enable isoman
sudo systemctl start isoman
```

---

## Production Considerations

### Security

1. **Run as non-root user**
   - Use dedicated user account
   - Limit file permissions

2. **Use HTTPS**
   - Deploy behind reverse proxy with SSL/TLS
   - Use Let's Encrypt for free certificates

3. **Firewall rules**
   ```bash
   # Only allow HTTP/HTTPS from reverse proxy
   sudo ufw allow from 127.0.0.1 to any port 8080
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

4. **File permissions**
   ```bash
   chmod 755 /opt/isoman/server
   chmod -R 755 /opt/isoman/ui/dist
   chmod 700 /opt/isoman/data
   ```

### Performance

1. **Worker count**
   - Default: 2 concurrent downloads
   - Increase for more parallelism: `WORKER_COUNT=4`
   - Consider network bandwidth and disk I/O

2. **Disk space**
   - Monitor available space
   - ISOs can be 1-5GB each
   - Set up disk space alerts

3. **Memory**
   - Minimum: 512MB RAM
   - Recommended: 1-2GB RAM
   - Downloads stream to disk, not memory

4. **Network**
   - Good internet connection for downloading ISOs
   - Consider bandwidth limits if on metered connection

### Monitoring

1. **Health check**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Logs**
   ```bash
   # Docker
   docker logs -f isoman

   # Systemd
   journalctl -u isoman -f
   ```

3. **Metrics to monitor**
   - Disk usage (`df -h`)
   - Active downloads
   - Failed downloads
   - Network bandwidth

### Backup

1. **Database**
   ```bash
   # Backup SQLite database
   cp /data/db/isos.db /backup/isos.db.$(date +%Y%m%d)
   ```

2. **ISO files**
   ```bash
   # Backup ISO directory (large!)
   rsync -av /data/isos/ /backup/isos/
   ```

3. **Docker volume**
   ```bash
   # Backup entire data volume
   docker run --rm \
     -v isoman-data:/data \
     -v $(pwd):/backup \
     alpine tar czf /backup/isoman-data-$(date +%Y%m%d).tar.gz /data
   ```

### Updates

1. **Docker**
   ```bash
   # Pull new image
   docker pull isoman:latest

   # Stop and remove old container
   docker stop isoman
   docker rm isoman

   # Start new container (data persists in volume)
   docker run -d -p 8080:8080 -v isoman-data:/data --name isoman isoman:latest
   ```

2. **Manual**
   ```bash
   # Stop service
   sudo systemctl stop isoman

   # Replace binary
   sudo cp server /opt/isoman/server

   # Replace UI
   sudo rm -rf /opt/isoman/ui/dist
   sudo cp -r dist /opt/isoman/ui/

   # Start service
   sudo systemctl start isoman
   ```

---

## Troubleshooting

### Container won't start

1. **Check logs**
   ```bash
   docker logs isoman
   ```

2. **Check port availability**
   ```bash
   netstat -tlnp | grep 8080
   ```

3. **Check permissions**
   ```bash
   docker run --rm -v isoman-data:/data alpine ls -la /data
   ```

### Downloads failing

1. **Check network connectivity**
   ```bash
   curl -I <download-url>
   ```

2. **Check disk space**
   ```bash
   df -h /data
   ```

3. **Check worker logs**
   - Look for error messages in container/service logs

### WebSocket not connecting

1. **Reverse proxy configuration**
   - Ensure WebSocket upgrade headers are set
   - Check `/ws` endpoint is proxied correctly

2. **Test WebSocket directly**
   ```bash
   wscat -c ws://localhost:8080/ws
   ```

### High memory usage

1. **Check active downloads**
   - Large ISOs stream to disk
   - Multiple concurrent downloads increase memory

2. **Reduce worker count**
   ```bash
   docker run -d -p 8080:8080 -v isoman-data:/data -e WORKER_COUNT=1 isoman
   ```

---

## Advanced Configuration

### Custom Data Directory

```bash
docker run -d \
  -p 8080:8080 \
  -v /mnt/storage:/custom-data \
  -e DATA_DIR=/custom-data \
  isoman
```

### External Database (Future)

Currently, ISOMan uses SQLite embedded in the data directory. For high availability, consider:
- Database replication (future feature)
- Read-only replicas (future feature)

### Load Balancing (Future)

For multiple instances:
- Shared storage (NFS, S3)
- Sticky sessions for WebSocket
- Database synchronization

---

## Best Practices

1. **Start small**
   - Begin with default settings
   - Monitor performance
   - Adjust worker count as needed

2. **Regular backups**
   - Backup database regularly
   - ISOs can be re-downloaded if lost

3. **Monitor disk space**
   - Set up alerts for low disk space
   - Clean up old/unused ISOs

4. **Use HTTPS**
   - Protect credentials (if added in future)
   - Secure WebSocket connections

5. **Keep up to date**
   - Regularly update Docker image
   - Monitor for security updates

---

## Support

For issues or questions:
- Check logs first
- Review this documentation
- Open an issue on GitHub
