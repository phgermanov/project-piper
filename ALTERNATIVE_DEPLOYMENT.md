# Alternative Deployment Options (Without Coolify)

If Coolify deployment is failing, here are alternative methods to deploy your docs.

## Option 1: Manual Deployment Script (Easiest)

This script handles everything: cloning, building, and setting up a reverse proxy.

### Prerequisites
- Ubuntu/Debian server with Docker installed
- Domain pointing to your server
- SSH access to your server

### Steps

1. SSH into your Hetzner server:
   ```bash
   ssh root@your-server-ip
   ```

2. Download and run the deployment script:
   ```bash
   curl -O https://raw.githubusercontent.com/phgermanov/project-piper/main/deploy-manual.sh
   chmod +x deploy-manual.sh
   ./deploy-manual.sh
   ```

3. Or manually copy the script from the repository and run it.

The script will:
- Clone/update the repository
- Build the Docker image
- Deploy the container
- Set up Caddy reverse proxy with automatic HTTPS

## Option 2: Docker Compose (More Control)

### Steps

1. SSH into your server and clone the repository:
   ```bash
   git clone https://github.com/phgermanov/project-piper.git
   cd project-piper
   ```

2. Deploy using Docker Compose:
   ```bash
   docker compose -f docker-compose.prod.yml up -d --build
   ```

3. Set up a reverse proxy (nginx or Caddy) to forward traffic to port 8080.

### Using Caddy as Reverse Proxy

```bash
# Install Caddy
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install caddy

# Configure Caddy
cat > /etc/caddy/Caddyfile <<EOF
piper.phgermanov.com {
    reverse_proxy localhost:8080
}
EOF

# Start Caddy
systemctl enable caddy
systemctl start caddy
```

### Using Nginx as Reverse Proxy

```bash
# Install nginx
apt install -y nginx certbot python3-certbot-nginx

# Configure nginx
cat > /etc/nginx/sites-available/piper-docs <<EOF
server {
    listen 80;
    server_name piper.phgermanov.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable site
ln -s /etc/nginx/sites-available/piper-docs /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx

# Get SSL certificate
certbot --nginx -d piper.phgermanov.com
```

## Option 3: Direct Docker Run

Quick and simple, but no automatic HTTPS:

```bash
# Clone repository
git clone https://github.com/phgermanov/project-piper.git
cd project-piper

# Build image
docker build \
  --build-arg DOCUSAURUS_URL=https://piper.phgermanov.com \
  --build-arg DOCUSAURUS_BASE_URL=/ \
  -t piper-docs \
  .

# Run container
docker run -d \
  --name piper-docs \
  --restart unless-stopped \
  -p 8080:80 \
  piper-docs

# Check it's running
docker ps
curl http://localhost:8080/health
```

Then set up a reverse proxy (Caddy or nginx) as shown in Option 2.

## Option 4: Fix Coolify

If you want to stick with Coolify, try these debugging steps:

### Check Coolify Server Logs

```bash
# SSH into your server
ssh root@your-server-ip

# Check Coolify logs
docker logs coolify -f

# Check Docker daemon
systemctl status docker

# Check disk space
df -h

# Check memory
free -h

# Try to manually build the image
cd /tmp
git clone https://github.com/phgermanov/project-piper.git
cd project-piper
docker build -t test-build .
```

### Restart Coolify

```bash
# Restart Coolify container
docker restart coolify

# Or if running as systemd service
systemctl restart coolify
```

### Check Coolify Proxy

In Coolify dashboard:
1. Go to Servers → Your server
2. Check Proxy status
3. If not running, restart it

### Try Coolify's Terminal Feature

If Coolify has a terminal or SSH feature:
1. Use it to access the server
2. Navigate to your deployment directory
3. Try to manually run the Docker build command
4. Check for error messages

## Updating the Deployment

For manual deployments, create an update script:

```bash
#!/bin/bash
cd /opt/project-piper
git pull origin main
docker build \
  --build-arg DOCUSAURUS_URL=https://piper.phgermanov.com \
  --build-arg DOCUSAURUS_BASE_URL=/ \
  -t piper-docs:latest \
  .
docker stop piper-docs
docker rm piper-docs
docker run -d \
  --name piper-docs \
  --restart unless-stopped \
  -p 8080:80 \
  piper-docs:latest
```

## Monitoring

Check container logs:
```bash
docker logs piper-docs -f
```

Check container status:
```bash
docker ps | grep piper-docs
```

Check health:
```bash
curl http://localhost:8080/health
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs piper-docs

# Check if port is already in use
netstat -tulpn | grep 8080

# Try a different port
docker run -d --name piper-docs -p 8081:80 piper-docs
```

### Build fails
```bash
# Check Docker disk space
docker system df

# Clean up old images
docker system prune -a

# Check build logs
docker build --no-cache -t piper-docs .
```

### Can't access from browser
```bash
# Check firewall
ufw status
ufw allow 80
ufw allow 443

# Check reverse proxy
systemctl status caddy
# or
systemctl status nginx

# Check DNS
dig piper.phgermanov.com
```

## Recommended Approach

**For production**: Use Option 1 (Manual Deployment Script) - it's simple, reliable, and includes automatic HTTPS.

**For development**: Use Option 3 (Direct Docker Run) - quick to test changes.

**To debug Coolify**: Follow Option 4 steps to identify the root cause.
