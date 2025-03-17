FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

COPY . .
RUN go build -o defdrive ./main.go

EXPOSE 8080
CMD ["./defdrive"]
