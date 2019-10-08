FROM golang:1.13-alpine AS build

WORKDIR /build

COPY cmd/main.go .

RUN apk add --no-cache git

RUN go get -u github.com/Azure/azure-sdk-for-go/...

RUN go get github.com/gorilla/handlers \
    && go get github.com/dimchansky/utfbom \
    && go get github.com/mitchellh/go-homedir \
    && go get golang.org/x/crypto/pkcs12

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app main.go

FROM alpine:3

WORKDIR /
COPY --from=build app .

CMD ["/app"]