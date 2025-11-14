# Coolify Deployment - Quick Start

This document provides quick instructions for deploying the Project Piper docs to Coolify.

## Quick Configuration (Recommended)

Use these settings in Coolify for a successful deployment:

### Repository Settings
- **Repository**: `https://github.com/phgermanov/project-piper`
- **Branch**: `claude/fix-coolify-deploy-docs-017zQxNhGvRoaWFpVNBThEZM`
- **Build Pack**: Docker

### Build Settings
- **Dockerfile Location**: `Dockerfile` (or leave empty)
- **Base Directory**: Leave empty (repository root)

### Build Arguments (Add in "Build" section)
```
DOCUSAURUS_URL=https://docs.phgermanov.com
DOCUSAURUS_BASE_URL=/
```

### Port Configuration
- **Port**: `80` (internal container port)

### Domain Configuration
- **Domain**: `docs.phgermanov.com`
- **HTTPS**: Enable (automatic Let's Encrypt)

### Health Check (Optional but recommended)
- **Health Check URL**: `/health`
- **Method**: `GET`
- **Expected Status**: `200`

## Default Credentials
- **Username**: `admin`
- **Password**: `changeme`

**⚠️ IMPORTANT**: Change these credentials before deploying to production!

## Common Errors and Solutions

### "Oops something is not okay" Error

This usually means:

1. **Dockerfile location is wrong**:
   - ✅ Correct: Dockerfile Location = `Dockerfile`, Base Directory = empty
   - ❌ Wrong: Dockerfile Location = `docs-site/Dockerfile`, Base Directory = empty

2. **Build arguments are missing**:
   - Add `DOCUSAURUS_URL` and `DOCUSAURUS_BASE_URL` in Build section

3. **Base directory misconfiguration**:
   - Use root directory (empty Base Directory) with the root-level Dockerfile

### Solution Steps

1. In Coolify, go to your resource settings
2. Set **Dockerfile Location** to just `Dockerfile`
3. Leave **Base Directory** empty
4. Add build arguments in the **Build** section
5. Save and redeploy

## Testing Locally

```bash
# From repository root
docker build \
  --build-arg DOCUSAURUS_URL=https://docs.phgermanov.com \
  --build-arg DOCUSAURUS_BASE_URL=/ \
  -t project-piper-docs .

# Run the container
docker run -p 8080:80 project-piper-docs

# Access at http://localhost:8080
# Default credentials: admin / changeme
```

## Full Documentation

For detailed documentation, see [docs-site/COOLIFY_DEPLOYMENT.md](docs-site/COOLIFY_DEPLOYMENT.md)

## Need Help?

1. Check Coolify build logs for detailed error messages
2. Verify all configuration settings match the Quick Configuration above
3. Ensure your domain DNS points to your Hetzner server
4. Make sure ports 80 and 443 are open on your firewall
