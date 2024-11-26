FROM golang:1.23-alpine3.20 AS builder

WORKDIR /workspace
COPY . .
RUN apk add --no-cache make && \
    make build

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /workspace/build/* ./
COPY --from=builder /workspace/config ./config

CMD ["/app/gitlab-mr-reviewer"]