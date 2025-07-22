# -----------------------------------------------
# ✅ Active: Multi-stage Production Build
# -----------------------------------------------

FROM golang:1.24-bullseye AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o app .

FROM debian:bullseye-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .

EXPOSE 8080

ENTRYPOINT ["/app/app"]

# -----------------------------------------------
# ✅ Alternative: Dev Mode (Uncomment to use)
# -----------------------------------------------

# FROM golang:1.24-bullseye
# WORKDIR /app
#
# COPY go.mod go.sum ./
# RUN go mod tidy
#
# COPY . .
#
# RUN go build -o app .
# EXPOSE 8080
# CMD ["./app"]
