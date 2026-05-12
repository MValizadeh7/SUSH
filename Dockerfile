FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod ./

COPY . .

RUN go build -o sush .


FROM alpine:3.23

RUN adduser -D -H -u 2000 sush

COPY --from=builder /build/sush /usr/local/bin/sush

EXPOSE 8080

USER sush

ENTRYPOINT [ "sush" ]