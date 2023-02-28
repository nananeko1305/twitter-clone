FROM golang:latest as builder
WORKDIR /app
COPY ./tweet_service/go.mod ./tweet_service/go.sum ./
RUN go mod download
COPY ./tweet_service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY /rbac/auth_model.conf/ .
COPY /tweet_service/policy.csv .
EXPOSE 8000
CMD ["./main"]
