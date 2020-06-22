FROM golang:1.14.4-alpine3.12 as builder

ARG Version

RUN apk --no-cache add git && \
    apk --update add alpine-sdk && \
    rm -rf /var/cache/apk/*

WORKDIR /sources
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -mod vendor -o tmass

FROM alpine:3.12

LABEL name="tmass"
LABEL description="Session Manager for Tmux"

RUN apk add --no-cache tmux

WORKDIR /

COPY --from=builder /sources/tmass tmass

ENTRYPOINT ["/tmass"]