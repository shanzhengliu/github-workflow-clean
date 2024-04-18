FROM golang:1.20.6-alpine3.18 AS BuildStage
WORKDIR /app
COPY . .
RUN apk --no-cache add upx
RUN go mod download && \
    go build -o /app/main .
RUN upx /app/main

FROM alpine:3.18
COPY --from=BuildStage /app/main /app/main
CMD ["/app/main"]

