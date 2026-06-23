# --- build stage ---
FROM golang:1.24-alpine AS build
WORKDIR /src

# Copy module files first to leverage Docker layer caching. There are no
# external dependencies, so this stays minimal.
COPY go.mod ./
COPY . .

# Build a static binary.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/aggregator .

# --- runtime stage ---
FROM alpine:3.20
WORKDIR /data
COPY --from=build /out/aggregator /usr/local/bin/aggregator

ENTRYPOINT ["aggregator"]
CMD ["--help"]
