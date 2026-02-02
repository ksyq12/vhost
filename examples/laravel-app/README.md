# Laravel Application Example

This example shows how to host a Laravel PHP framework application with proper routing and PHP-FPM configuration.

## Scenario

You have a Laravel application that needs:
- PHP-FPM for processing PHP files
- URL rewriting for Laravel's routing system
- Public directory as the web root
- SSL for production deployment

## Prerequisites

```bash
# Install PHP and required extensions
sudo apt install php8.2-fpm php8.2-cli php8.2-mysql php8.2-xml php8.2-mbstring

# Install Composer
sudo apt install composer

# Verify PHP-FPM is running
sudo systemctl status php8.2-fpm
```

## Setup Steps

### 1. Create Laravel Project

```bash
# Create project directory
cd /var/www
composer create-project laravel/laravel myapp

# Set permissions
sudo chown -R $USER:www-data /var/www/myapp
sudo chmod -R 775 /var/www/myapp/storage /var/www/myapp/bootstrap/cache
```

### 2. Configure Environment

```bash
cd /var/www/myapp
cp .env.example .env
php artisan key:generate

# Edit .env for your database settings
nano .env
```

### 3. Add Virtual Host

```bash
sudo vhost add myapp.test \
    --type laravel \
    --root /var/www/myapp \
    --php 8.2
```

**Important:** Specify the Laravel root directory, not the public subdirectory. vhost automatically routes to `public/`.

### 4. Update Hosts File (for local development)

```bash
echo "127.0.0.1 myapp.test" | sudo tee -a /etc/hosts
```

### 5. Run Migrations

```bash
cd /var/www/myapp
php artisan migrate
```

### 6. Test

```bash
# Test HTTP
curl http://myapp.test

# Or visit in browser
firefox http://myapp.test
```

## Common Tasks

### Running Artisan Commands

```bash
cd /var/www/myapp
php artisan route:list
php artisan cache:clear
php artisan config:cache
```

### Enabling SSL for Production

```bash
sudo vhost ssl install myapp.test --email admin@myapp.test

# Update .env
APP_URL=https://myapp.test
```

## Directory Permissions

Laravel requires specific permissions:

```bash
# Fix ownership
sudo chown -R $USER:www-data /var/www/myapp

# Fix directory permissions
sudo find /var/www/myapp -type d -exec chmod 755 {} \;
sudo find /var/www/myapp -type f -exec chmod 644 {} \;

# Set proper permissions for storage and cache
sudo chmod -R 775 /var/www/myapp/storage
sudo chmod -R 775 /var/www/myapp/bootstrap/cache
```

## Troubleshooting

### 502 Bad Gateway

Check PHP-FPM status:
```bash
sudo systemctl status php8.2-fpm
sudo systemctl restart php8.2-fpm
```

### Permission Errors

Check error logs:
```bash
vhost logs myapp.test --error

# Common fix
sudo chown -R $USER:www-data /var/www/myapp
sudo chmod -R 775 /var/www/myapp/storage
```

### Route Not Found (404)

Ensure your vhost is properly configured:
```bash
vhost show myapp.test

# Verify nginx config syntax
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

## Cleaning Up

```bash
# Remove virtual host
sudo vhost remove myapp.test --force

# Remove application (optional)
sudo rm -rf /var/www/myapp
```
