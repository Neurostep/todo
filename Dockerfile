FROM golang:1.13.4 AS builder
WORKDIR /src
COPY Makefile go.mod go.sum ./
RUN make install
COPY . .
RUN make build-release

FROM alpine:3.10 AS runtime
RUN adduser -h /app -s /bin/sh -D app && chown -R app: /app

COPY --from=builder /src/build/todo /app/

USER app
WORKDIR /app
ENTRYPOINT ["./todo"]