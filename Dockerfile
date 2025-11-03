FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine AS builder

WORKDIR /app

ADD . .

ARG TARGETARCH TARGETOS

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o node-manager cmd/daemon/main.go

FROM alpine:latest as release
COPY --from=builder /app/node-manager /usr/local/bin/node-manager
CMD ["node-manager"]
