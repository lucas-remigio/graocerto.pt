#!/bin/bash

# HTTP/2 Testing Script
# Usage: ./test-http2.sh your-domain.com

DOMAIN=${1:-"your-domain.com"}

echo "🚀 Testing HTTP/2 support for $DOMAIN"
echo "================================================"

# Test 1: Check HTTP/2 support with curl
echo "1. Testing HTTP/2 with curl..."
curl -I --http2 -s "https://$DOMAIN" | head -1

# Test 2: Check protocol version
echo -e "\n2. Checking protocol version..."
curl -w "HTTP Version: %{http_version}\nResponse Code: %{response_code}\n" -o /dev/null -s "https://$DOMAIN"

# Test 3: Check if server pushes resources (if configured)
echo -e "\n3. Testing multiplexing capabilities..."
curl --http2 -w "Total time: %{time_total}s\nConnect time: %{time_connect}s\n" -o /dev/null -s "https://$DOMAIN"

# Test 4: Check security headers
echo -e "\n4. Checking security headers..."
curl -I -s "https://$DOMAIN" | grep -i "strict-transport-security\|x-frame-options\|x-content-type-options"

echo -e "\n✅ HTTP/2 test completed!"
echo "💡 You can also test in browser DevTools > Network tab > Protocol column"
