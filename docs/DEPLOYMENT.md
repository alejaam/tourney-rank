# Deployment Guide: Dokploy + Cloudflare Tunnel on Raspberry Pi

This guide explains how to deploy TourneyRank on a Raspberry Pi using Dokploy and expose it via Cloudflare Tunnel.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                        INTERNET                               │
│                            │                                  │
│                    Cloudflare (DNS + SSL)                     │
│                            │                                  │
│                   ┌────────▼────────┐                        │
│                   │ Cloudflare Tunnel│                        │
│                   │   (cloudflared)  │                        │
│                   └────────┬────────┘                        │
└────────────────────────────┼─────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────┐
│                     RASPBERRY PI                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                      Docker                              │ │
│  │  ┌─────────────┐  ┌──────────┐  ┌──────────┐           │ │
│  │  │ TourneyRank │  │ PostgreSQL│  │  Redis   │           │ │
│  │  │    :8080    │  │  :5432   │  │  :6379   │           │ │
│  │  └──────┬──────┘  └────┬─────┘  └────┬─────┘           │ │
│  │         └───────────────┴─────────────┘                 │ │
│  │                  tourneyrank-net                        │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                     Dokploy                              │ │
│  │              (manages deployments)                       │ │
│  └─────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Raspberry Pi 4 (4GB+ RAM recommended) or Pi 5
- Raspberry Pi OS (64-bit recommended for better performance)
- Docker and Docker Compose installed
- Dokploy installed and running
- Cloudflare account with a domain
- Domain configured in Cloudflare DNS

## Step 1: Prepare the Raspberry Pi

### Install Docker

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
# Log out and back in for group changes
```

### Install Dokploy

Follow [Dokploy installation guide](https://docs.dokploy.com/get-started/introduction) or:

```bash
curl -sSL https://dokploy.com/install.sh | sh
```

Access Dokploy at `http://<raspberry-pi-ip>:3000`

## Step 2: Setup Cloudflare Tunnel

### Create Tunnel in Cloudflare Dashboard

1. Go to Cloudflare Dashboard → Zero Trust → Access → Tunnels
2. Click "Create a tunnel"
3. Name it (e.g., `raspberrypi-tunnel`)
4. Copy the tunnel token

### Install cloudflared on Raspberry Pi

```bash
# Download cloudflared for ARM64
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64 -o cloudflared
chmod +x cloudflared
sudo mv cloudflared /usr/local/bin/

# Install as service
sudo cloudflared service install <YOUR_TUNNEL_TOKEN>

# Verify it's running
sudo systemctl status cloudflared
```

### Configure Tunnel Routes

In Cloudflare Dashboard → Tunnels → Your Tunnel → Public Hostname:

| Subdomain | Domain | Service |
|-----------|--------|---------|
| api | yourdomain.com | http://localhost:8080 |
| tourneyrank | yourdomain.com | http://localhost:8080 |

This routes `api.yourdomain.com` and `tourneyrank.yourdomain.com` to your TourneyRank backend.

## Step 3: Deploy with Dokploy

### Option A: Deploy from Git (Recommended)

1. In Dokploy, create a new Project
2. Add a new Service → Application
3. Choose "Git" as source
4. Connect your GitHub repository
5. Set branch to `main`
6. Dokploy will detect the Dockerfile automatically

### Option B: Deploy with Docker Compose

1. In Dokploy, create a new Project
2. Add a new Service → Compose
3. Paste the `docker-compose.yml` content
4. Configure environment variables

### Environment Variables in Dokploy

Set these in Dokploy's Environment section:

```env
ENVIRONMENT=production
LOG_LEVEL=info
HTTP_PORT=8080
DATABASE_URL=postgresql://tourneyrank:YOUR_STRONG_PASSWORD@postgres:5432/tourneyrank?sslmode=disable
REDIS_URL=redis://redis:6379
POSTGRES_PASSWORD=YOUR_STRONG_PASSWORD
JWT_SECRET=YOUR_RANDOM_SECRET_HERE
```

Generate strong secrets:
```bash
# Generate random password
openssl rand -base64 32

# Generate JWT secret
openssl rand -hex 32
```

## Step 4: Configure Health Checks

Dokploy uses health checks to know when the app is ready. The TourneyRank backend exposes:

- `/healthz` - Liveness probe (is the process alive?)
- `/readyz` - Readiness probe (can it accept traffic?)

Configure in Dokploy:
- Health Check Path: `/healthz`
- Health Check Interval: 30s
- Health Check Timeout: 5s

## Step 5: SSL/TLS Configuration

**You don't need to configure SSL on the Raspberry Pi!**

Cloudflare Tunnel handles SSL termination:
- User → HTTPS → Cloudflare → Tunnel → HTTP → Raspberry Pi

In Cloudflare:
1. SSL/TLS → Overview → Set to "Full" or "Full (strict)"
2. SSL/TLS → Edge Certificates → Enable "Always Use HTTPS"

## Step 6: Resource Optimization for Raspberry Pi

### Docker Resource Limits

Add to `docker-compose.yml` for production:

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          memory: 256M

  postgres:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M

  redis:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 128M
```

### PostgreSQL Tuning for Low Memory

Create `postgresql.conf` or set environment variables:

```yaml
postgres:
  environment:
    POSTGRES_USER: tourneyrank
    POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    POSTGRES_DB: tourneyrank
  command: >
    postgres
    -c shared_buffers=128MB
    -c effective_cache_size=256MB
    -c maintenance_work_mem=32MB
    -c work_mem=4MB
    -c max_connections=50
```

## Step 7: Monitoring

### Check Application Health

```bash
# From Raspberry Pi
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/debug/info

# From external (via Cloudflare)
curl https://api.yourdomain.com/healthz
```

### View Logs

```bash
# Via Docker
docker logs tourneyrank-api -f

# Via Dokploy
# Use Dokploy's built-in log viewer
```

### Resource Usage

```bash
# Check container stats
docker stats

# Check Pi resources
htop
df -h  # Disk usage (important for SD cards!)
```

## Step 8: Backups

### Database Backup Script

Create `/home/pi/backup-tourneyrank.sh`:

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/home/pi/backups
mkdir -p $BACKUP_DIR

# Backup PostgreSQL
docker exec tourneyrank-db pg_dump -U tourneyrank tourneyrank > $BACKUP_DIR/db_$DATE.sql
gzip $BACKUP_DIR/db_$DATE.sql

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "Backup completed: db_$DATE.sql.gz"
```

Add to crontab:
```bash
crontab -e
# Add:
0 2 * * * /home/pi/backup-tourneyrank.sh >> /home/pi/backup.log 2>&1
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker logs tourneyrank-api

# Check if ports are in use
sudo lsof -i :8080

# Rebuild image
docker-compose build --no-cache backend
```

### Cloudflare Tunnel not connecting

```bash
# Check cloudflared status
sudo systemctl status cloudflared

# Check tunnel logs
sudo journalctl -u cloudflared -f

# Restart tunnel
sudo systemctl restart cloudflared
```

### Database connection issues

```bash
# Check PostgreSQL is healthy
docker exec tourneyrank-db pg_isready -U tourneyrank

# Check network
docker network inspect tourneyrank-net
```

### Out of memory (OOM)

```bash
# Check memory usage
free -h

# Check for OOM kills
dmesg | grep -i "out of memory"

# Reduce container limits in docker-compose.yml
```

### SD Card wear (for Pi with SD card)

```bash
# Check SD card health (if using SD)
sudo apt install smartmontools
sudo smartctl -a /dev/mmcblk0

# Consider using an SSD for better reliability
```

## Production Checklist

- [ ] Strong passwords set for PostgreSQL and JWT_SECRET
- [ ] ENVIRONMENT=production
- [ ] LOG_LEVEL=info (not debug)
- [ ] Cloudflare SSL set to Full or Full (strict)
- [ ] Health checks configured in Dokploy
- [ ] Backup script scheduled
- [ ] Resource limits set for containers
- [ ] PostgreSQL tuned for Pi memory
- [ ] Firewall configured (only allow Cloudflare IPs if exposed directly)
- [ ] Docker logs rotation configured

## Useful Commands

```bash
# Restart all services
docker-compose restart

# View all containers
docker ps -a

# Enter container shell
docker exec -it tourneyrank-api sh

# Run database migrations
docker exec tourneyrank-api ./tourneyrank migrate up

# Check disk space
df -h

# Monitor in real-time
watch -n 2 'docker stats --no-stream'
```

## Next Steps

1. Set up monitoring with Prometheus/Grafana (optional)
2. Configure webhook notifications for deployments
3. Set up CI/CD in GitHub Actions
4. Add more API endpoints as needed

---

**Questions?** Check the main README or open an issue on GitHub.
