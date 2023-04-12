FROM golang:1.20-alpine AS builder
WORKDIR /handlers-service
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
WORKDIR /handlers-service/cmd/app
RUN CGO_ENABLED=0 go build -o /bin/app
WORKDIR /handlers-service

FROM alpine:latest
RUN apk --update add ca-certificates
COPY --from=builder /bin/app /bin/app
COPY --from=builder /handlers-service/deploy_handlers_service/.env /deploy_handlers_service/.env

CMD [ "/bin/app" ]