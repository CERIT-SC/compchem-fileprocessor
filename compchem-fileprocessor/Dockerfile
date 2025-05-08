FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o invenio-file-processor .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/invenio-file-processor .

ENTRYPOINT ["./invenio-file-processor"]
