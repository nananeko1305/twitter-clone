FROM golang:latest as builder
WORKDIR /app
COPY ./user_service/go.mod ./user_service/go.sum ./
RUN go mod download
COPY ./user_service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /root/
COPY /rbac/auth_model.conf/ .
COPY /user_service/policy.csv .
COPY --from=builder /app/main .
EXPOSE 8000
CMD ["./main"]
