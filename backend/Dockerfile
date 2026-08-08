# --- dev: hot reload with air ---
FROM golang:1.26-alpine AS dev
WORKDIR /app
# Pinned, not @latest: a floating tool version breaks the image without warning.
RUN go install github.com/air-verse/air@v1.67.4
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["air", "-c", ".air.toml"]

# --- build: compile static binary ---
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

# --- prod: minimal runtime image ---
FROM alpine:3.21 AS prod
WORKDIR /app
COPY --from=build /bin/api /app/api
RUN mkdir -p /app/uploads
EXPOSE 8080
CMD ["/app/api"]
