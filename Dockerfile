FROM golang:1.17.0-bullseye

RUN go version

ENV GOPATH=/

COPY ./ ./

RUN apt-get update
RUN apt-get -y install postgresql-client
RUN chmod +x wait-for-postgres.sh


RUN go mod download

RUN go build -o tg-sota-feedback ./main.go

CMD ["./tg-sota-feedback"]