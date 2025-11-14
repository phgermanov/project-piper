# Multi-stage build for Docusaurus documentation
# This Dockerfile is at the repository root for easier Coolify deployment
# Stage 1: Build the static site
FROM node:20-alpine AS builder

WORKDIR /app

# Build arguments for Docusaurus configuration
ARG DOCUSAURUS_URL=https://docs.phgermanov.com
ARG DOCUSAURUS_BASE_URL=/

# Set environment variables for the build
ENV DOCUSAURUS_URL=${DOCUSAURUS_URL}
ENV DOCUSAURUS_BASE_URL=${DOCUSAURUS_BASE_URL}

# Copy package files from docs-site directory
COPY docs-site/package*.json ./

# Install dependencies (including devDependencies needed for build)
RUN npm ci

# Copy source files from docs-site directory
COPY docs-site/ .

# Build the static site
RUN npm run build

# Stage 2: Serve with nginx and basic auth
FROM nginx:alpine

# Install apache2-utils for htpasswd
RUN apk add --no-cache apache2-utils

# Copy the built site from builder stage
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx configuration from docs-site
COPY docs-site/nginx.conf /etc/nginx/conf.d/default.conf

# Create directory for auth files
RUN mkdir -p /etc/nginx/auth

# Create a default .htpasswd file (should be replaced via environment or mounted volume)
# Default credentials: admin/changeme
RUN htpasswd -cb /etc/nginx/auth/.htpasswd admin changeme

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
  CMD wget --quiet --tries=1 --spider http://localhost/health || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
