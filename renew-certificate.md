# Stop the frontend container temporarily

docker stop lucas-frontend-1

# Renew certificates

sudo certbot renew --force-renewal

# Start the container again

docker start lucas-frontend-1
