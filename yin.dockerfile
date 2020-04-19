# Making build for project
FROM golang:alpine AS goBuilder
WORKDIR /go/src/yin
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -installsuffix cgo -o yin

# Running project with the build
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
EXPOSE 3000:3000
COPY --from=goBuilder /go/src/yin/yin /go/src/yin/production.env /go/src/yin/local.env ./
# RUN mkdir temp
# RUN cat > temp.txt
# RUN ls && pwd
RUN echo "....................................................................*......................................."
CMD ["./yin","-env","production"]  

