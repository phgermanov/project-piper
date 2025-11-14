# Deploying Project Piper Documentation to Coolify

This guide explains how to deploy the Project Piper documentation site to Coolify with basic authentication.

## Prerequisites

- Access to Coolify instance at `admin.phgermanov.com`
- Hetzner server configured with Coolify
- Domain or subdomain configured (e.g., `docs.phgermanov.com`)

## Features

- Dockerized Docusaurus documentation site
- Built-in basic authentication with nginx
- Health check endpoint
- Optimized multi-stage build
- Gzip compression
- Security headers

## Deployment Steps

### 1. Create New Project in Coolify

1. Log in to Coolify at `https://admin.phgermanov.com`
2. Navigate to **Projects** > **New Project**
3. Click **Add New Resource** > **Public Repository**

### 2. Configure Git Repository

- **Repository URL**: `https://github.com/phgermanov/project-piper`
- **Branch**: `claude/hetzner-coolify-setup-01PcVbDXbBu4iREgujvfDzHB` (or your preferred branch)
- **Build Pack**: Docker
- **Dockerfile Location**: `docs-site/Dockerfile`
- **Docker Context**: `docs-site`

### 3. Configure Environment Variables

Add the following environment variables in Coolify:

#### Build Arguments (for Docker build)

```bash
DOCUSAURUS_URL=https://docs.phgermanov.com
DOCUSAURUS_BASE_URL=/
```

These can be set in the **Build** section of your Coolify resource.

### 4. Configure Authentication

#### Option A: Use Default Credentials (Not Recommended for Production)

The default credentials are built into the Docker image:
- **Username**: `admin`
- **Password**: `changeme`

#### Option B: Create Custom Credentials (Recommended)

1. Generate a .htpasswd file locally:

```bash
# Install apache2-utils (if not already installed)
# Ubuntu/Debian
sudo apt-get install apache2-utils

# Alpine
apk add apache2-utils

# macOS
brew install httpd

# Create .htpasswd file with your username
htpasswd -c .htpasswd yourusername
# Enter your password when prompted
```

2. In Coolify, add the .htpasswd content as a **Secret File**:
   - Go to your resource settings
   - Navigate to **Secrets** > **Add Secret**
   - Name: `HTPASSWD_FILE`
   - Content: Paste the contents of your .htpasswd file

3. Mount the secret in the Dockerfile by modifying the `docker-compose.yml` or add a volume mount in Coolify:
   - Path: `/etc/nginx/auth/.htpasswd`
   - Source: Your secret file

Alternatively, you can create a new .htpasswd file in the repository:

```bash
cd docs-site
htpasswd -c .htpasswd yourusername
```

Then update the Dockerfile to copy this file instead of creating a default one.

### 5. Configure Domain

1. In Coolify, go to **Domains** section of your resource
2. Add your domain: `docs.phgermanov.com`
3. Enable **HTTPS** (Coolify will automatically provision Let's Encrypt SSL)
4. Set **Port**: `80` (internal container port)

### 6. Configure Health Check

Coolify can use the built-in health check:

- **Health Check URL**: `/health`
- **Health Check Method**: `GET`
- **Expected Status Code**: `200`

### 7. Deploy

1. Click **Deploy** in Coolify
2. Monitor the build logs
3. Once deployed, access your documentation at `https://docs.phgermanov.com`

## Testing Locally

Before deploying to Coolify, you can test locally:

```bash
cd docs-site

# Build the Docker image
docker build -t project-piper-docs .

# Run the container
docker run -p 8080:80 project-piper-docs

# Or use docker-compose
docker-compose up
```

Access at: `http://localhost:8080`

Default credentials: `admin` / `changeme`

## Creating Custom Authentication

To change the authentication credentials before deployment:

1. Create a new .htpasswd file:

```bash
cd docs-site
htpasswd -c .htpasswd newusername
```

2. Update the Dockerfile to copy your .htpasswd file:

```dockerfile
# Replace this line in the Dockerfile:
RUN htpasswd -cb /etc/nginx/auth/.htpasswd admin changeme

# With:
COPY .htpasswd /etc/nginx/auth/.htpasswd
```

3. Make sure to add `.htpasswd` to `.gitignore` if it contains production credentials:

```bash
echo ".htpasswd" >> .gitignore
```

## Advanced Configuration

### Multiple Users

To add multiple users to .htpasswd:

```bash
# Create first user
htpasswd -c .htpasswd user1

# Add additional users (without -c flag)
htpasswd .htpasswd user2
htpasswd .htpasswd user3
```

### Disable Authentication for Specific Paths

Edit `nginx.conf` and add `auth_basic off;` to specific location blocks:

```nginx
location /public {
    auth_basic off;
    try_files $uri $uri/ /index.html;
}
```

### Custom nginx Configuration

You can modify `docs-site/nginx.conf` to:
- Add custom headers
- Configure additional security settings
- Set up redirects
- Add rate limiting

## Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DOCUSAURUS_URL` | Full URL where docs will be hosted | `https://docs.phgermanov.com` | Yes |
| `DOCUSAURUS_BASE_URL` | Base path for the site | `/` | Yes |

## Troubleshooting

### Build Fails

1. Check Coolify build logs for errors
2. Verify the Docker context path is set to `docs-site`
3. Ensure Node.js version compatibility (requires Node 20+)

### Authentication Not Working

1. Verify .htpasswd file is properly mounted
2. Check nginx logs in Coolify
3. Ensure credentials are correct (try default: admin/changeme)

### Site Not Loading Correctly

1. Verify `DOCUSAURUS_URL` and `DOCUSAURUS_BASE_URL` are set correctly
2. Check browser console for 404 errors
3. Verify domain is properly configured in Coolify

### SSL Issues

1. Ensure domain DNS is pointing to your Hetzner server
2. Check Coolify SSL certificate provisioning status
3. Verify port 80 and 443 are open on your firewall

## Security Recommendations

1. **Always change default credentials** before deploying to production
2. Use strong passwords (generate with: `openssl rand -base64 32`)
3. Keep the .htpasswd file out of version control
4. Enable HTTPS/SSL (automatic with Coolify + Let's Encrypt)
5. Regularly update the base Docker images
6. Monitor access logs for suspicious activity

## Updating the Documentation

After pushing changes to your repository:

1. Coolify can auto-deploy on git push (configure webhooks)
2. Or manually trigger a redeploy in Coolify dashboard
3. Monitor build logs to ensure successful deployment

## Support

For issues with:
- **Coolify**: Check Coolify documentation or admin.phgermanov.com
- **Documentation content**: Open an issue in the GitHub repository
- **Build errors**: Review the Dockerfile and build logs

## Additional Resources

- [Coolify Documentation](https://coolify.io/docs)
- [Docusaurus Documentation](https://docusaurus.io/)
- [nginx Documentation](https://nginx.org/en/docs/)
- [Apache htpasswd Documentation](https://httpd.apache.org/docs/current/programs/htpasswd.html)
