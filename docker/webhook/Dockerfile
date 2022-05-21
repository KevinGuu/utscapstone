# dev env
FROM golang:1.18.0-alpine as dev-env

WORKDIR /app

# prod env
FROM dev-env as build-env

COPY go.mod /go.sum /app/
RUN go mod download

COPY . /app/
RUN CGO_ENABLED=0 go build -o /webhook

FROM alpine:latest as runtime

COPY --from=build-env /webhook /usr/local/bin/webhook
RUN chmod +x /usr/local/bin/webhook

ENTRYPOINT ["webhook"]