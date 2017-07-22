FROM golang:1.8.3

# Create app directory
RUN mkdir -p /go/src/github.com/lempiy/gochat
WORKDIR /go/src/github.com/lempiy/gochat

# Install app dependencies
RUN go get github.com/beego/bee

COPY . /go/src/github.com/lempiy/gochat

EXPOSE 8001
CMD [ "bee", "run" ]