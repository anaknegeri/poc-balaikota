# Use official Node.js 20 Alpine image
FROM node:20-alpine AS base

# Install pnpm globally
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Install dependencies only when needed
FROM base AS deps
# Copy package files
COPY package.json pnpm-lock.yaml* ./
# Install dependencies with pnpm
RUN pnpm install --frozen-lockfile --shamefully-hoist

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
# Ensure node_modules is properly linked
RUN ln -sf /app/node_modules/.bin /usr/local/bin

# Set environment variable for production build
ENV SKIP_TYPE_CHECK=true
ENV NODE_ENV=production

# Accept build arguments for environment variables
ARG VITE_API_BASE_URL
ARG VITE_DASHBOARD_REFRESH_INTERVAL
ARG VITE_ALERTS_REFRESH_INTERVAL
ARG VITE_CHARTS_REFRESH_INTERVAL
ARG VITE_CHART_TRENDS_LIMIT
ARG VITE_CHART_ALERTS_LIMIT
ARG VITE_DEFAULT_THEME
ARG VITE_ANIMATION_DURATION
ARG VITE_ENABLE_AUTO_REFRESH
ARG VITE_ENABLE_IMAGE_PREVIEW

# Set environment variables from build arguments
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_DASHBOARD_REFRESH_INTERVAL=$VITE_DASHBOARD_REFRESH_INTERVAL
ENV VITE_ALERTS_REFRESH_INTERVAL=$VITE_ALERTS_REFRESH_INTERVAL
ENV VITE_CHARTS_REFRESH_INTERVAL=$VITE_CHARTS_REFRESH_INTERVAL
ENV VITE_CHART_TRENDS_LIMIT=$VITE_CHART_TRENDS_LIMIT
ENV VITE_CHART_ALERTS_LIMIT=$VITE_CHART_ALERTS_LIMIT
ENV VITE_DEFAULT_THEME=$VITE_DEFAULT_THEME
ENV VITE_ANIMATION_DURATION=$VITE_ANIMATION_DURATION
ENV VITE_ENABLE_AUTO_REFRESH=$VITE_ENABLE_AUTO_REFRESH
ENV VITE_ENABLE_IMAGE_PREVIEW=$VITE_ENABLE_IMAGE_PREVIEW

# Build the application
RUN pnpm run build:docker

# Production image, copy all the files and run the app
FROM nginx:alpine AS runner

# Set working directory
WORKDIR /usr/share/nginx/html

# Remove default nginx static assets
RUN rm -rf ./*

# Copy built application from builder stage
COPY --from=builder /app/dist .

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:80 || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
