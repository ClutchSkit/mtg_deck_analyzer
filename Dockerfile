FROM golang:1.27 AS builder
WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 go build -o mtg-analyzer
FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/mtg-analyzer /app/mtg-analyzer
CMD ["/app/mtg-analyzer"]