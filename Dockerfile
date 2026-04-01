FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY backend/go.mod backend/go.sum* ./
RUN go mod download

COPY backend/ .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /build/project-works-server ./cmd/server

# ---

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/project-works-server /app/project-works-server
COPY --from=builder /build/migrations /app/migrations

EXPOSE 8088

ENTRYPOINT ["/app/project-works-server"]
