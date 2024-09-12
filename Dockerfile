FROM golang:1.21

RUN mkdir /goapp

WORKDIR  /goapp

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o /docker-go-restaurent

RUN go get -d -v 


EXPOSE 8000

CMD ["/docker-go-restaurent"]