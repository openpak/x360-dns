# Multi-stage, cross-compiled: the build stage runs on the runner's native arch and Go
# emits the target arch, so linux/arm64 builds need no QEMU. x360-common comes from vendor/,
# so the build needs nothing outside this repository.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/x360-dns ./cmd/x360-dns

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/x360-dns /usr/local/bin/x360-dns
USER app
EXPOSE 53/udp 53/tcp
ENTRYPOINT ["/usr/local/bin/x360-dns"]
