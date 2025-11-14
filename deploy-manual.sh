#!/bin/bash
# Manual deployment script for Hetzner server (without Coolify)
# Run this on your Hetzner server

set -e

echo "🚀 Deploying Project Piper Docs..."

# Configuration
REPO_URL="https://github.com/phgermanov/project-piper.git"
BRANCH="main"
DOMAIN="piper.phgermanov.com"
CONTAINER_NAME="piper-docs"
IMAGE_NAME="piper-docs:latest"

# Clone or update repository
if [ -d "/opt/project-piper" ]; then
    echo "📦 Updating repository..."
    cd /opt/project-piper
    git fetch origin
    git checkout $BRANCH
    git pull origin $BRANCH
else
    echo "📦 Cloning repository..."
    git clone -b $BRANCH $REPO_URL /opt/project-piper
    cd /opt/project-piper
fi

# Build Docker image
echo "🔨 Building Docker image..."
docker build \
    --build-arg DOCUSAURUS_URL=https://$DOMAIN \
    --build-arg DOCUSAURUS_BASE_URL=/ \
    -t $IMAGE_NAME \
    .

# Stop and remove old container if it exists
echo "🛑 Stopping old container..."
docker stop $CONTAINER_NAME 2>/dev/null || true
docker rm $CONTAINER_NAME 2>/dev/null || true

# Run new container
echo "▶️  Starting new container..."
docker run -d \
    --name $CONTAINER_NAME \
    --restart unless-stopped \
    -p 8080:80 \
    $IMAGE_NAME

# Set up Caddy reverse proxy (if not already configured)
if ! command -v caddy &> /dev/null; then
    echo "📝 Setting up Caddy reverse proxy..."
    apt-get update
    apt-get install -y debian-keyring debian-archive-keyring apt-transport-https
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
    apt-get update
    apt-get install -y caddy
fi

# Configure Caddy
echo "⚙️  Configuring Caddy..."
cat > /etc/caddy/Caddyfile <<EOF
$DOMAIN {
    reverse_proxy localhost:8080
}
EOF

# Reload Caddy
systemctl reload caddy

echo "✅ Deployment complete!"
echo "🌐 Your documentation is available at: https://$DOMAIN"
echo ""
echo "📊 Container status:"
docker ps | grep $CONTAINER_NAME
