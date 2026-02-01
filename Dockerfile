# Build the Go backend
FROM golang:1.23-bullseye AS builder

# Allow the Go tool to fetch the required 1.24 toolchain automatically
ENV GOTOOLCHAIN=auto

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backend ./

# Runtime image with a prebuilt headless Chrome
FROM chromedp/headless-shell:latest

ENV CHROME_PATH=/headless-shell/headless-shell
# Ensure Chromium can write runtime files
ENV XDG_RUNTIME_DIR=/tmp/chrome-runtime
ENV CHROME_LOG_FILE=/tmp/chrome.log
RUN mkdir -p /tmp/chrome-runtime /tmp/chrome-crash && chmod 700 /tmp/chrome-runtime /tmp/chrome-crash

WORKDIR /app

# Create working dirs for static files
RUN mkdir -p /app/uploads /app/compressed

COPY --from=builder /src/contracts /app/contracts
COPY --from=builder /src/images /app/images
COPY --from=builder /src/index.html /app/index.html
COPY --from=builder /out/backend /usr/local/bin/backend

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/backend"]
