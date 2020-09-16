# stage 0: compile go program
FROM golang:1.15-alpine
RUN apk add --no-cache make && mkdir -p /tmp/rdr-emailer
WORKDIR /tmp/rdr-emailer
ADD Makefile .
ADD main.go .
ADD go.mod .
ADD go.sum .
RUN make build_linux_amd64 && mv rdr-emailer.linux_amd64 rdr-emailer

# stage 1: build image for the rdr-emailer container
FROM alpine:latest as rdr-emailer
WORKDIR /
COPY --from=0 /tmp/rdr-emailer/rdr-emailer .

CMD ["/rdr-emailer","-h"]
