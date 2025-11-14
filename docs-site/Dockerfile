# Multi-stage build for Docusaurus documentation
# Stage 1: Build the static site
FROM node:20-alpine AS builder

WORKDIR /app

# Build arguments for Docusaurus configuration
ARG DOCUSAURUS_URL=https://docs.phgermanov.com
ARG DOCUSAURUS_BASE_URL=/

# Set environment variables for the build
ENV DOCUSAURUS_URL=${DOCUSAURUS_URL}
ENV DOCUSAURUS_BASE_URL=${DOCUSAURUS_BASE_URL}

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source files
COPY . .

# Build the static site
RUN npm run build

# Stage 2: Serve with nginx and basic auth
FROM nginx:alpine

# Install apache2-utils for htpasswd
RUN apk add --no-cache apache2-utils

# Copy the built site from builder stage
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Create directory for auth files
RUN mkdir -p /etc/nginx/auth

# Create a default .htpasswd file (should be replaced via environment or mounted volume)
# Default credentials: admin/changeme
RUN htpasswd -cb /etc/nginx/auth/.htpasswd admin changeme

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
