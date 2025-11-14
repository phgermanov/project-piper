# Website

This website is built using [Docusaurus](https://docusaurus.io/), a modern static website generator.

## Installation

```bash
yarn
```

## Local Development

```bash
yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```bash
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Deployment

### GitHub Pages

Using SSH:

```bash
USE_SSH=true yarn deploy
```

Not using SSH:

```bash
GIT_USER=<Your GitHub username> yarn deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.

### Coolify / Docker Deployment

This documentation site can be deployed to Coolify with built-in authentication:

1. **See [COOLIFY_DEPLOYMENT.md](./COOLIFY_DEPLOYMENT.md)** for complete deployment instructions
2. **Quick start with Docker**:

```bash
# Build the image
docker build -t project-piper-docs .

# Run locally
docker run -p 8080:80 project-piper-docs

# Or use docker-compose
docker-compose up
```

3. **Setup authentication**:

```bash
# Create .htpasswd file for authentication
./create-auth.sh
```

Access at: `http://localhost:8080`
Default credentials: `admin` / `changeme` (change for production!)

For detailed Coolify setup instructions, custom authentication, and production deployment, see [COOLIFY_DEPLOYMENT.md](./COOLIFY_DEPLOYMENT.md).
