# Production stage — expects pre-built output from `npm run build`
# Build locally: cd frontend && npm install && PUBLIC_API_URL=http://localhost:8001 npm run build
FROM node:20-alpine
WORKDIR /app

# Copy built app and production deps
COPY build ./build
COPY package*.json ./
RUN npm ci --omit=dev

# Create non-root user for security
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 sveltekit
USER sveltekit

# Set environment
ENV NODE_ENV=production
ENV PORT=3000
ENV ORIGIN=http://localhost:3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

EXPOSE 3000

CMD ["node", "build"]
