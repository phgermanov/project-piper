#!/bin/bash
# Script to create .htpasswd file for nginx basic authentication
# Usage: ./create-auth.sh

set -e

echo "========================================="
echo "Project Piper Docs - Authentication Setup"
echo "========================================="
echo ""

# Check if htpasswd is installed
if ! command -v htpasswd &> /dev/null; then
    echo "Error: htpasswd is not installed."
    echo ""
    echo "Please install apache2-utils:"
    echo "  Ubuntu/Debian: sudo apt-get install apache2-utils"
    echo "  Alpine: apk add apache2-utils"
    echo "  macOS: brew install httpd"
    echo ""
    exit 1
fi

# Get username
read -p "Enter username: " username

if [ -z "$username" ]; then
    echo "Error: Username cannot be empty"
    exit 1
fi

# Check if .htpasswd already exists
if [ -f .htpasswd ]; then
    read -p ".htpasswd file already exists. Do you want to add another user? (y/n): " add_user
    if [ "$add_user" = "y" ] || [ "$add_user" = "Y" ]; then
        htpasswd .htpasswd "$username"
    else
        read -p "Do you want to recreate .htpasswd file? (y/n): " recreate
        if [ "$recreate" = "y" ] || [ "$recreate" = "Y" ]; then
            htpasswd -c .htpasswd "$username"
        else
            echo "Aborted."
            exit 0
        fi
    fi
else
    htpasswd -c .htpasswd "$username"
fi

echo ""
echo "========================================="
echo "Success!"
echo "========================================="
echo ""
echo ".htpasswd file created successfully!"
echo ""
echo "IMPORTANT SECURITY NOTES:"
echo "1. Keep this file secure and never commit it to version control"
echo "2. Add .htpasswd to .gitignore if not already present"
echo "3. Use this file when deploying to Coolify"
echo ""
echo "To add more users, run: htpasswd .htpasswd <username>"
echo ""

# Check if .htpasswd is in .gitignore
if ! grep -q "^\.htpasswd$" .gitignore 2>/dev/null; then
    echo "Adding .htpasswd to .gitignore..."
    echo ".htpasswd" >> .gitignore
    echo "✓ Added to .gitignore"
fi
