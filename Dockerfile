FROM golang:1.18.1-alpine3.15 as builder
WORKDIR /app/
RUN apk add build-base
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /go/bin/app main.go


FROM alpine:3.15
WORKDIR /app/
USER 1000
COPY --from=builder /go/bin/app .
CMD ["./app"]
