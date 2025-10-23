FROM docker.io/library/golang:1.25.3-alpine as builder

RUN apk add --no-cache git ca-certificates build-base
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o external-dns-opnsense .

FROM docker.io/library/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/external-dns-opnsense .
EXPOSE 8888
CMD ["./external-dns-opnsense"]