FROM golang:latest as builder
WORKDIR /app
COPY ./follow_service/go.mod ./follow_service/go.sum ./
RUN go mod download
COPY ./follow_service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY /rbac/auth_model.conf/ .
COPY /follow_service/policy.csv .
EXPOSE 8000
CMD ["./main"]
