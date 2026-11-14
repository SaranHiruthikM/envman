# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache deps separately from source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build both binaries static, stripped.
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X github.com/SaranHiruthikM/envman/cmd/envman/commands.Version=${VERSION}" \
    -o /out/envman-server ./cmd/server
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/envman-migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot AS server
COPY --from=build /out/envman-server /usr/local/bin/envman-server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/envman-server"]

FROM gcr.io/distroless/static-debian12:nonroot AS migrate
COPY --from=build /out/envman-migrate /usr/local/bin/envman-migrate
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/envman-migrate"]
