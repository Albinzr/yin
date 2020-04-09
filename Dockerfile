#Get go image
#install packages
#build go
#copy executable build to location
#delete other files
#create volume for temp files
#run executable

FROM golang
WORKDIR /go/src/yin
COPY . .
RUN go mod download
RUN go build
CMD ["./yin"]