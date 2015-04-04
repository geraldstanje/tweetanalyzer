FROM golang

# Add directories
ADD flickgo /go/src/github.com/manki/flickgo/

RUN ["go", "get", "github.com/geraldstanje/tweetanalyzer" ]

WORKDIR /go/src/github.com/geraldstanje/tweetanalyzer

# Add directories and files
ADD config.xml /go/src/github.com/geraldstanje/tweetanalyzer/config.xml

ENTRYPOINT go build && ./tweetanalyzer

EXPOSE 8080