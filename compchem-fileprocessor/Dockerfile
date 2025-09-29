FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o compchem-fileprocessor .

FROM scratch

WORKDIR /app

COPY migrations /app/migrations
COPY --from=builder /app/compchem-fileprocessor .

ENTRYPOINT ["./compchem-fileprocessor"]
