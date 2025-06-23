FROM node:18-alpine

WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install production dependencies only
RUN npm ci --only=production && npm cache clean --force

# Copy application code
COPY . .

# Create non-root user for security
RUN addgroup -g 1001 -S nodejs && \
    adduser -S socketuser -u 1001 && \
    chown -R socketuser:nodejs /app

# Switch to non-root user
USER socketuser

EXPOSE 8090

CMD ["npm", "start"]