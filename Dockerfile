FROM golang:1.13-alpine AS build

WORKDIR /build

COPY cmd/main.go .

RUN apk add --no-cache git \
    && go get -u github.com/Azure/azure-sdk-for-go/... \
    && go get github.com/gorilla/handlers

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app main.go

FROM alpine:3

WORKDIR /
COPY --from=build app .

CMD ["/app"]