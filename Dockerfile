FROM golang:latest

RUN go get gitlab.com/hokiegeek/life
RUN mkdir -p /go/src/gitlab.com/hokiegeek/biologist
ADD . /go/src/gitlab.com/hokiegeek/biologist
RUN go install gitlab.com/hokiegeek/biologist/...

ENTRYPOINT ["biologistd"]
