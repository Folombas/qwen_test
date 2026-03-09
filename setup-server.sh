#!/bin/bash
# Go Quiz — Server Setup Script
# Автоматическая настройка Ubuntu сервера для Docker

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    log_error "Please run as root (sudo ./setup-server.sh)"
    exit 1
fi

log_step "=== Go Quiz Server Setup ==="
echo ""

# 1. Update system
log_step "1. Updating system..."
apt update && apt upgrade -y
log_info "System updated"

# 2. Install essential packages
log_step "2. Installing essential packages..."
apt install -y \
    curl \
    git \
    wget \
    ufw \
    fail2ban \
    htop \
    nano \
    vim \
    sqlite3 \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release
log_info "Essential packages installed"

# 3. Install Docker
log_step "3. Installing Docker..."

# Add Docker's official GPG key
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

# Add Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add user to docker group
if [ -n "$SUDO_USER" ] && [ "$SUDO_USER" != "root" ]; then
    usermod -aG docker $SUDO_USER
    log_info "User $SUDO_USER added to docker group"
fi

# Enable Docker service
systemctl enable docker
systemctl start docker

log_info "Docker installed successfully"

# 4. Install Docker Compose (standalone)
log_step "4. Installing Docker Compose..."
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
mkdir -p $DOCKER_CONFIG/cli-plugins/
curl -SL https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose
chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
log_info "Docker Compose installed"

# 5. Configure UFW firewall
log_step "5. Configuring firewall (UFW)..."
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw --force enable
log_info "Firewall configured"

# 6. Configure fail2ban (SSH protection)
log_step "6. Configuring fail2ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF

systemctl enable fail2ban
systemctl start fail2ban
log_info "fail2ban configured"

# 7. Configure sysctl for better performance
log_step "7. Optimizing system..."
cat >> /etc/sysctl.conf << 'EOF'

# Docker optimizations
net.ipv4.ip_forward=1
net.core.somaxconn=65535
vm.max_map_count=262144
EOF

sysctl -p
log_info "System optimized"

# 8. Create application directory
log_step "8. Creating application directory..."
mkdir -p /opt/qwen_test
chown -R $SUDO_USER:$SUDO_USER /opt/qwen_test 2>/dev/null || chown -R root:root /opt/qwen_test
log_info "Directory created: /opt/qwen_test"

# 9. Display server info
log_step "=== Server Information ==="
echo ""
echo "Docker version:"
docker --version
echo ""
echo "Docker Compose version:"
docker compose version
echo ""
echo "UFW status:"
ufw status verbose
echo ""
echo "Fail2ban status:"
systemctl is-active fail2ban
echo ""

# 10. Next steps
log_step "=== Next Steps ==="
echo ""
echo "1. Clone your repository:"
echo "   cd /opt/qwen_test"
echo "   git clone https://github.com/Folombas/qwen_test.git ."
echo ""
echo "2. Configure environment:"
echo "   cp .env.example .env"
echo "   nano .env  # Change JWT_SECRET and DOMAIN_NAME"
echo ""
echo "3. Deploy with Docker:"
echo "   docker compose up -d"
echo ""
echo "4. Check status:"
echo "   docker compose ps"
echo "   curl http://localhost:8080/api/stats"
echo ""
echo "5. Setup SSL (after domain is configured):"
echo "   ./deploy.sh ssl"
echo ""
log_info "Server setup completed! 🎉"
