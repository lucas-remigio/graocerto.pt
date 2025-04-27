# Use a Node base image
FROM node:18-alpine AS build

# Set the working directory
WORKDIR /app

# Define build arguments
# Vite required the ENV variables to be passed as build arguments
ARG VITE_BACKEND_URL
ARG VITE_SOCKETS_URL

# Make them available as environment variables during build
ENV VITE_BACKEND_URL=$VITE_BACKEND_URL
ENV VITE_SOCKETS_URL=$VITE_SOCKETS_URL

# Install pnpm globally
RUN npm install -g pnpm

# Copy only package.json and pnpm-lock.yaml first to leverage Docker caching
COPY package.json pnpm-lock.yaml ./

# Install dependencies and ensure optional dependencies are skipped
RUN pnpm install 

# Copy the rest of the application
COPY . .

# Build the app
RUN pnpm build

# Serve the build with Nginx
FROM nginx:stable
COPY nginx.template.conf /etc/nginx/conf.d/default.conf.template
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80 443

# Use environment variable substitution for ALL variables before starting nginx
CMD ["/bin/bash", "-c", "envsubst '$BACKEND_URL $BACKEND_PORT $SOCKETS_URL $SOCKETS_PORT $FRONTEND_DOMAIN $FRONTEND_URL' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]