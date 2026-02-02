# Reverse Proxy Example

This example demonstrates using vhost to set up a reverse proxy to backend applications like Node.js, Python, or Go services.

## Scenario

You have a backend application running on a local port, and you want to:
- Expose it through a domain name
- Add SSL termination
- Handle WebSocket connections
- Forward proper headers

## Use Cases

- Node.js/Express API servers
- Python Flask/Django applications
- Go HTTP servers
- Docker containers exposing HTTP ports

## Setup Steps

### 1. Start Your Backend Application

Example Node.js app:

```bash
# Create a simple Express app
mkdir -p ~/apps/api
cd ~/apps/api

npm init -y
npm install express

cat > index.js << 'EOF'
const express = require('express');
const app = express();

app.get('/', (req, res) => {
    res.json({
        message: 'Hello from backend!',
        headers: req.headers
    });
});

app.listen(3000, () => {
    console.log('Backend running on port 3000');
});
EOF

# Start the application
node index.js
```

Verify it's working:
```bash
curl http://localhost:3000
```

### 2. Add Reverse Proxy Virtual Host

```bash
sudo vhost add api.example.com \
    --type proxy \
    --proxy http://localhost:3000
```

### 3. Update Hosts File (for local testing)

```bash
echo "127.0.0.1 api.example.com" | sudo tee -a /etc/hosts
```

### 4. Test the Proxy

```bash
# Test HTTP
curl http://api.example.com

# Should return the same as:
curl http://localhost:3000
```

### 5. Enable SSL

```bash
sudo vhost ssl install api.example.com --email admin@example.com

# Now accessible via HTTPS
curl https://api.example.com
```

## Multiple Backend Services

You can proxy multiple services on different subdomains:

```bash
# API service
sudo vhost add api.example.com --type proxy --proxy http://localhost:3000

# Admin panel
sudo vhost add admin.example.com --type proxy --proxy http://localhost:4000

# WebSocket service
sudo vhost add ws.example.com --type proxy --proxy http://localhost:5000
```

## Docker Container Example

### 1. Run Container

```bash
docker run -d \
    --name my-api \
    -p 8080:8080 \
    my-api-image
```

### 2. Proxy to Container

```bash
sudo vhost add api.example.com \
    --type proxy \
    --proxy http://localhost:8080
```

## WebSocket Support

The vhost proxy configuration automatically handles WebSocket upgrades. Test with:

```javascript
// Client-side code
const ws = new WebSocket('wss://api.example.com');

ws.onopen = () => {
    console.log('Connected');
    ws.send('Hello server!');
};

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};
```

## Running Backend as a Service

Create a systemd service:

```bash
sudo nano /etc/systemd/system/my-api.service
```

```ini
[Unit]
Description=My API Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/home/user/apps/api
ExecStart=/usr/bin/node /home/user/apps/api/index.js
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable my-api
sudo systemctl start my-api
```

## Monitoring and Debugging

### Check Backend Status

```bash
# Verify backend is running
curl http://localhost:3000

# Check listening ports
sudo netstat -tlnp | grep :3000
```

### View Proxy Logs

```bash
# Access log
vhost logs api.example.com --access

# Error log
vhost logs api.example.com --error

# Follow logs in real-time
vhost logs api.example.com -f
```

## Troubleshooting

### 502 Bad Gateway

Backend is not running or not accessible:
```bash
# Check if backend is up
curl http://localhost:3000

# Restart backend
sudo systemctl restart my-api
```

### 504 Gateway Timeout

Request took too long. Edit the nginx config for longer timeout:
```bash
vhost edit api.example.com
```

Add to location block:
```nginx
proxy_read_timeout 300s;
proxy_connect_timeout 300s;
```

## Cleaning Up

```bash
# Remove virtual host
sudo vhost remove api.example.com --force

# Stop backend service
sudo systemctl stop my-api
sudo systemctl disable my-api
```
