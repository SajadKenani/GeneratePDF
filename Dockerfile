# Build the Go backend
FROM golang:1.23-bullseye AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backend ./

FROM chromedp/headless-shell:latest

ENV CHROME_PATH=/headless-shell/headless-shell
ENV XDG_RUNTIME_DIR=/tmp/chrome-runtime
ENV CHROME_LOG_FILE=/tmp/chrome.log

RUN mkdir -p /tmp/chrome-runtime /tmp/chrome-crash && chmod 700 /tmp/chrome-runtime /tmp/chrome-crash

RUN apt-get update && apt-get install -y --no-install-recommends \
    pdftk \
    poppler-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN mkdir -p /app/uploads /app/compressed

COPY --from=builder /src/pdf /app/pdf
COPY --from=builder /src/images /app/images
COPY --from=builder /src/index.html /app/index.html
COPY --from=builder /out/backend /usr/local/bin/backend

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/backend"]