FROM golang:alpine
WORKDIR /go/src/github.com/
COPY ./pd-analyze .
EXPOSE 8080
ENTRYPOINT ["/pd-analyze"]