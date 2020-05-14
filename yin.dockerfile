# Making build for project
FROM golang:alpine AS goBuilder
WORKDIR /go/src/yin
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -installsuffix cgo -o yin

# Running project with the build
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
EXPOSE 1000:1000
COPY --from=goBuilder /go/src/yin/yin /go/src/yin/production.env /go/src/yin/local.env ./
CMD ["./yin","-env","production"]  
