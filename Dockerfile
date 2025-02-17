FROM golang:1.23-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o tinycache cmd/tinycache/main.go

FROM scratch as runner

COPY --from=builder /app/tinycache /tinycache

EXPOSE 8080

CMD ["./tinycache"]