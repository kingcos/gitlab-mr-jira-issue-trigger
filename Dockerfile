FROM golang:1.6.2-onbuild
MAINTAINER kingcos

RUN ln -s /go/bin/app /go/bin/gitlab-ci-mr-jira-issue-trigger
EXPOSE 8989
