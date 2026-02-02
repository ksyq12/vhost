# WordPress Example

Complete guide for hosting a WordPress site with optimized caching, security hardening, and SSL support.

## Scenario

You want to run a WordPress blog or website with:
- PHP-FPM for performance
- Static file caching
- Security hardening (block xmlrpc.php, wp-config.php)
- Permalink support
- SSL encryption

## Prerequisites

```bash
# Install PHP and MySQL
sudo apt install php8.2-fpm php8.2-mysql php8.2-curl php8.2-gd php8.2-xml php8.2-mbstring php8.2-zip
sudo apt install mysql-server

# Install WP-CLI (optional but recommended)
curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar
chmod +x wp-cli.phar
sudo mv wp-cli.phar /usr/local/bin/wp
```

## Setup Steps

### 1. Create Database

```bash
sudo mysql -u root -p

# In MySQL prompt:
CREATE DATABASE wordpress;
CREATE USER 'wpuser'@'localhost' IDENTIFIED BY 'secure_password';
GRANT ALL PRIVILEGES ON wordpress.* TO 'wpuser'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 2. Download WordPress

```bash
# Create directory
sudo mkdir -p /var/www/myblog
cd /var/www/myblog

# Download and extract
sudo wp core download --allow-root

# Set ownership
sudo chown -R $USER:www-data /var/www/myblog
```

### 3. Configure WordPress

```bash
# Create wp-config.php
sudo wp config create \
    --dbname=wordpress \
    --dbuser=wpuser \
    --dbpass=secure_password \
    --allow-root
```

### 4. Add Virtual Host

```bash
sudo vhost add myblog.com \
    --type wordpress \
    --root /var/www/myblog \
    --php 8.2
```

### 5. Complete WordPress Installation

Visit `http://myblog.com/wp-admin/install.php` in your browser, or use WP-CLI:

```bash
sudo wp core install \
    --url=http://myblog.com \
    --title="My Blog" \
    --admin_user=admin \
    --admin_password=admin_password \
    --admin_email=admin@myblog.com \
    --allow-root
```

### 6. Enable SSL

```bash
sudo vhost ssl install myblog.com --email admin@myblog.com

# Update WordPress URL
sudo wp option update home 'https://myblog.com' --allow-root
sudo wp option update siteurl 'https://myblog.com' --allow-root
```

## WordPress Security

### 1. Disable File Editing

Add to `wp-config.php`:
```php
define('DISALLOW_FILE_EDIT', true);
```

### 2. Keep Everything Updated

```bash
sudo wp core update --allow-root
sudo wp plugin update --all --allow-root
sudo wp theme update --all --allow-root
```

### 3. Set Correct Permissions

```bash
# Set ownership
sudo chown -R www-data:www-data /var/www/myblog

# Set directory permissions
sudo find /var/www/myblog -type d -exec chmod 755 {} \;

# Set file permissions
sudo find /var/www/myblog -type f -exec chmod 644 {} \;

# Secure wp-config.php
sudo chmod 600 /var/www/myblog/wp-config.php
```

## Common Tasks

### Update WordPress

```bash
sudo wp core update --allow-root
sudo wp core update-db --allow-root
```

### Manage Plugins

```bash
# List plugins
sudo wp plugin list --allow-root

# Update all plugins
sudo wp plugin update --all --allow-root
```

### Database Backup

```bash
# Backup database
sudo wp db export /var/backups/myblog-$(date +%Y%m%d).sql --allow-root
```

## Troubleshooting

### White Screen of Death

Enable debugging in wp-config.php:
```php
define('WP_DEBUG', true);
define('WP_DEBUG_LOG', true);
define('WP_DEBUG_DISPLAY', false);
```

Check the log:
```bash
tail -f /var/www/myblog/wp-content/debug.log
```

### Upload Errors

Edit PHP-FPM config:
```bash
sudo nano /etc/php/8.2/fpm/php.ini

# Increase these values
upload_max_filesize = 64M
post_max_size = 64M
memory_limit = 256M

# Restart PHP-FPM
sudo systemctl restart php8.2-fpm
```

## Cleaning Up

```bash
# Remove virtual host
sudo vhost remove myblog.com --force

# Drop database
sudo mysql -u root -p -e "DROP DATABASE wordpress;"

# Remove files (optional)
sudo rm -rf /var/www/myblog
```
