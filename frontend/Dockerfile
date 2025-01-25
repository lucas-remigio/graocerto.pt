# Use a Node base image
FROM node:18-alpine AS build

# Set the working directory
WORKDIR /app

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
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]