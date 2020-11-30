FROM alpine:3.9 as builder

RUN apk add --no-cache curl
COPY ./build/VERSION VERSION
RUN version=$(cat VERSION) && curl -L "https://github.com/kubernetes/kompose/releases/download/v${version}/kompose-linux-amd64" -o kompose

FROM alpine:3.9

COPY --from=builder /kompose /usr/bin/kompose
RUN chmod +x /usr/bin/kompose
