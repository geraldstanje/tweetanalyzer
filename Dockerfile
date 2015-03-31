FROM golang

# Add directories
# the flickgo hat local modifications, thats why i need to copy it
ADD flickgo /go/src/github.com/manki/flickgo/
ADD . /go/src/github.com/geraldstanje/tweetanalyzer

RUN ["go", "get", "github.com/geraldstanje/tweetanalyzer" ]

WORKDIR /go/src/github.com/geraldstanje/tweetanalyzer

ENTRYPOINT go build && ./tweetanalyzer

EXPOSE 8080