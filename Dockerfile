FROM google/golang-runtime

WORKDIR /gopath/src/app

ENTRYPOINT ["/gopath/bin/app"]
