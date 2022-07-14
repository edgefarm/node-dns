FROM golang:1.18 AS builder
WORKDIR /code
COPY . .
RUN make

# FROM gcr.io/distroless/base
FROM alpine:3.14
COPY --from=builder /code/bin/node-dns /node-dns
ENTRYPOINT ["/node-dns"]
