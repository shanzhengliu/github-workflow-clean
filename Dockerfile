FROM --platform=$BUILDPLATFORM golang:1.20.6-alpine3.18 AS BuildStage
WORKDIR /app
COPY . .
RUN apk --no-cache add upx
RUN go mod download
RUN echo "Building for $TARGETPLATFORM" && \
    if [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
        export GOOS=linux GOARCH=amd64; \
    elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
        export GOOS=linux GOARCH=arm64; \
    fi && \
    go mod download && \
    go build -o /app/main .
RUN upx /app/main

FROM --platform=$TARGETPLATFORM alpine:3.18
COPY --from=BuildStage /app/main /app/main
CMD ["/app/main"]

