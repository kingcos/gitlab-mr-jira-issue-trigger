FROM golang:1.11.1
MAINTAINER github.com/kingcos <2821836721v@gmail.com>

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY . /go/src/app

RUN go get -v -d
RUN go install -v

RUN go build gitlab-mr-jira-issue-trigger.go

RUN ln -s /go/src/app/gitlab-mr-jira-issue-trigger /go/bin/gitlab-mr-jira-issue-trigger

EXPOSE 9090
