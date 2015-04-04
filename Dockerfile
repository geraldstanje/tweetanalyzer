FROM golang

# Add directories
#ADD twitterstream /go/src/github.com/darkhelmet/twitterstream/
ADD flickgo /go/src/github.com/manki/flickgo/
ADD . /go/src/github.com/geraldstanje/tweetanalyzer

RUN ["go", "get", "github.com/geraldstanje/tweetanalyzer" ]

WORKDIR /go/src/github.com/geraldstanje/tweetanalyzer

ENTRYPOINT go build && ./tweetanalyzer

EXPOSE 8080