# Static Site Example

This example demonstrates hosting a simple static HTML website with optional SSL.

## Scenario

You have a static website with HTML, CSS, and JavaScript files that you want to serve over HTTP and HTTPS.

## Directory Structure

```
/var/www/example.com/
├── index.html
├── css/
│   └── style.css
├── js/
│   └── script.js
└── images/
    └── logo.png
```

## Setup Steps

### 1. Create Document Root

```bash
sudo mkdir -p /var/www/example.com
sudo chown -R $USER:$USER /var/www/example.com
```

### 2. Add Sample Content

```bash
cat > /var/www/example.com/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Example Site</title>
    <link rel="stylesheet" href="/css/style.css">
</head>
<body>
    <h1>Welcome to Example.com</h1>
    <p>This is a static site hosted with vhost.</p>
    <script src="/js/script.js"></script>
</body>
</html>
EOF

mkdir -p /var/www/example.com/css /var/www/example.com/js
echo "body { font-family: Arial; }" > /var/www/example.com/css/style.css
echo "console.log('Static site loaded');" > /var/www/example.com/js/script.js
```

### 3. Add Virtual Host (HTTP only)

```bash
sudo vhost add example.com \
    --type static \
    --root /var/www/example.com
```

### 4. Verify

```bash
# Check if vhost is added
vhost list

# Test in browser
curl http://example.com
```

### 5. Enable SSL (Optional)

```bash
# Install SSL certificate
sudo vhost ssl install example.com --email admin@example.com

# Verify SSL is working
curl https://example.com
```

## Troubleshooting

### Site not loading
```bash
# Check web server status
sudo systemctl status nginx

# Check logs
vhost logs example.com --error
```

### Permission denied errors
```bash
# Fix ownership
sudo chown -R www-data:www-data /var/www/example.com

# Fix permissions
sudo find /var/www/example.com -type d -exec chmod 755 {} \;
sudo find /var/www/example.com -type f -exec chmod 644 {} \;
```

## Cleaning Up

```bash
# Remove virtual host
sudo vhost remove example.com --force

# Remove files (optional)
sudo rm -rf /var/www/example.com
```
