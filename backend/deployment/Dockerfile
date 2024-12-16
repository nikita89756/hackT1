FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

# RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN CGO_ENABLED=0 go build -o main ./cmd/hack/main.go

EXPOSE 8088

# RUN goose -dir deployment/migrations postgres "postgres://postgres:postgres@postgres:5432/testdb" up

CMD ["./main"]