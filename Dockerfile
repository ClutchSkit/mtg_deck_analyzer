FROM golang:1.27
WORKDIR /app
COPY . ./
RUN go build -o mtg-analyzer
CMD ["/app/mtg-analyzer"]