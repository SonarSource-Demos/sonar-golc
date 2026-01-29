# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -tags=golc -ldflags "-X main.version1=1.0.9" -o golc golc.go && \
    CGO_ENABLED=0 go build -trimpath -tags=resultsall -o ResultsAll ResultsAll.go

# Run stage: minimal Alpine for fewest vulnerabilities and small size
FROM alpine:3.19

RUN apk add --no-cache ca-certificates && \
    adduser -D -g "" appuser

WORKDIR /app

COPY --from=builder /build/golc /build/ResultsAll ./
COPY --from=builder /build/dist ./dist
COPY --from=builder /build/imgs ./imgs
COPY docker-entrypoint.sh ./

RUN chmod +x docker-entrypoint.sh && chown -R appuser:appuser /app

USER appuser

# GOLC_DEVOPS: which platform to analyze (e.g. Github, Gitlab, BitBucket, File); must match a key in config.json
ENV GOLC_CONFIG_FILE=/config/config.json

VOLUME ["/config", "/data"]

WORKDIR /data

# Entrypoint copies dist/imgs into /data then runs golc then ResultsAll (all from /data)

EXPOSE 8092

ENTRYPOINT ["/app/docker-entrypoint.sh"]
