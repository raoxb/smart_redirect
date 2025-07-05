#!/bin/bash

# SSL Certificate Generation Script
# Usage: ./scripts/ssl.sh [domain] [environment]

DOMAIN=${1:-localhost}
ENVIRONMENT=${2:-dev}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SSL_DIR="$PROJECT_DIR/ssl"

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Create SSL directory
mkdir -p "$SSL_DIR"

if [[ "$ENVIRONMENT" == "dev" ]]; then
    log "Generating self-signed SSL certificate for development..."
    
    # Generate private key
    openssl genrsa -out "$SSL_DIR/key.pem" 2048
    
    # Generate certificate signing request
    openssl req -new -key "$SSL_DIR/key.pem" -out "$SSL_DIR/cert.csr" -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"
    
    # Generate self-signed certificate
    openssl x509 -req -in "$SSL_DIR/cert.csr" -signkey "$SSL_DIR/key.pem" -out "$SSL_DIR/cert.pem" -days 365
    
    # Create CA bundle (for dev, just copy the cert)
    cp "$SSL_DIR/cert.pem" "$SSL_DIR/ca-bundle.pem"
    
    # Clean up CSR
    rm "$SSL_DIR/cert.csr"
    
    success "Self-signed SSL certificate generated successfully"
    log "Certificate: $SSL_DIR/cert.pem"
    log "Private key: $SSL_DIR/key.pem"
    log "CA bundle: $SSL_DIR/ca-bundle.pem"
    
else
    log "Production SSL certificate setup..."
    warning "For production, you should use certificates from a trusted CA like Let's Encrypt"
    
    cat << 'EOF'
For production SSL certificates, consider using Let's Encrypt with certbot:

1. Install certbot:
   sudo apt-get install certbot python3-certbot-nginx

2. Generate certificate:
   sudo certbot certonly --standalone -d yourdomain.com

3. Copy certificates to SSL directory:
   sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /path/to/ssl/cert.pem
   sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /path/to/ssl/key.pem
   sudo cp /etc/letsencrypt/live/yourdomain.com/chain.pem /path/to/ssl/ca-bundle.pem

4. Set up auto-renewal:
   sudo crontab -e
   Add: 0 12 * * * /usr/bin/certbot renew --quiet

EOF
fi

# Set proper permissions
chmod 600 "$SSL_DIR/key.pem"
chmod 644 "$SSL_DIR/cert.pem"
chmod 644 "$SSL_DIR/ca-bundle.pem" 2>/dev/null || true

success "SSL setup completed"