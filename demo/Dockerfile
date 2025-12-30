# Build stage - compile Go to WASM using TinyGo
FROM tinygo/tinygo:0.33.0 AS builder

WORKDIR /app
COPY main.go .

# Build WASM binary
RUN tinygo build -o main.wasm -target wasm main.go

# Runtime stage - serve with nginx
FROM nginx:alpine

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy web files
COPY index.html /usr/share/nginx/html/
COPY main.js /usr/share/nginx/html/

# Copy TinyGo's wasm_exec.js
COPY --from=builder /usr/local/tinygo/targets/wasm_exec.js /usr/share/nginx/html/

# Copy built WASM module
COPY --from=builder /app/main.wasm /usr/share/nginx/html/

EXPOSE 80
