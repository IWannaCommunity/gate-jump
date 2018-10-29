# Compile stage
FROM golang:1.10.1-alpine3.7 AS build-env
RUN apk update && apk add git 
WORKDIR $GOPATH/src
RUN git clone https://github.com/klazen108/gate-jump
WORKDIR $GOPATH/src/gate-jump/src/api
RUN go get
WORKDIR $GOPATH/src/gate-jump
RUN mkdir /app
RUN go build -v -o /app/gate-jump -i ./src/api/
 
# Run stage
FROM alpine:3.7
EXPOSE 8080
WORKDIR /
COPY --from=build-env /app/gate-jump /
ADD config /config
CMD ["/gate-jump"]