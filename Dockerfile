FROM golang:1.16-alpine
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=0 /app/main ./cmd/
# COPY --from=0 /app/bootstrap.json /

CMD ["./cmd/main"]