FROM golang:1.23.0 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/app /app/app

EXPOSE 8000
USER nonroot:nonroot

CMD ["/app/app"]
