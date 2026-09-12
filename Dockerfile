FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /common-svr .

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /common-svr /app/common-svr
COPY config /app/config
USER app
EXPOSE 8080
ENTRYPOINT ["/app/common-svr"]
