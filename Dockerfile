# build stage
FROM golang:alpine AS build-env

ENV GOPATH=/go
RUN mkdir -p /go/src/github.com/allanhung/sccc && mkdir -p /go/bin
ADD . /go/src/github.com/allanhung/sccc
RUN cd /go/src/github.com/allanhung/sccc && go build -o /go/bin/sccc
 
# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/bin/sccc /app
CMD ./app
