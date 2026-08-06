# Use a Node base image
FROM node:18-alpine AS build

# Set the working directory
WORKDIR /app

# Define build arguments
# Vite required the ENV variables to be passed as build arguments
ARG BACKEND_URL
ARG SOCKETS_URL
ARG BACKEND_PORT
ARG SOCKETS_PORT
ARG IS_PRODUCTION
ARG PUBLIC_GOOGLE_CLIENT_ID
ARG PUBLIC_VAPID_KEY
ARG ENABLE_WEBSOCKETS=false
ARG TELEGRAM_BOT_USERNAME=GraoCertoBot

# Make them available as environment variables during build
ENV VITE_BACKEND_URL=$BACKEND_URL
ENV VITE_SOCKETS_URL=$SOCKETS_URL
ENV VITE_BACKEND_PORT=$BACKEND_PORT
ENV VITE_SOCKETS_PORT=$SOCKETS_PORT
ENV VITE_IS_PRODUCTION=$IS_PRODUCTION
ENV VITE_PUBLIC_GOOGLE_CLIENT_ID=$PUBLIC_GOOGLE_CLIENT_ID
ENV VITE_PUBLIC_VAPID_KEY=$PUBLIC_VAPID_KEY
ENV VITE_ENABLE_WEBSOCKETS=$ENABLE_WEBSOCKETS
ENV VITE_TELEGRAM_BOT_USERNAME=$TELEGRAM_BOT_USERNAME

# Install pnpm globally
RUN npm install -g pnpm

# Copy only package.json and pnpm-lock.yaml first to leverage Docker caching
COPY package.json pnpm-lock.yaml ./

# Install dependencies (production only for smaller image)
RUN pnpm install --frozen-lockfile

# Copy the rest of the application
COPY . .

# Sync SvelteKit and build the app
RUN pnpm run prepare && pnpm build

# Production stage with Nginx
FROM nginxinc/nginx-unprivileged:alpine
# Remove default nginx config to prevent port 80 conflicts
RUN rm /etc/nginx/conf.d/default.conf
# Copy the file to the templates directory for environment variable substitution
COPY nginx.conf /etc/nginx/templates/production.conf.template
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]