FROM golang:1.12.0-alpine3.9
RUN apk --no-cache add git
RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go mod download

RUN go build -o main .

EXPOSE 9123
CMD ["/app/main"]