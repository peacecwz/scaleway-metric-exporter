FROM golang:1.20-alpine AS builder

RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER appuser

COPY --from=builder /app/main .

EXPOSE 9706
CMD ["./main"]
