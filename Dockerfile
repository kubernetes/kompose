# Alpine Builder
FROM alpine as builder

RUN apk add --no-cache curl
COPY ./build/VERSION VERSION
RUN \
  version=$(cat VERSION) && \
  ARCH=$(uname -m | sed 's/armv7l/arm/g' | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g') && \
  curl -L \
    "https://github.com/kubernetes/kompose/releases/download/v${version}/kompose-linux-${ARCH}" \
    -o kompose && \
  chmod +x kompose

# Runtime
FROM alpine

COPY --from=builder /kompose /usr/bin/kompose
