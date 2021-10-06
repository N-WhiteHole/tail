FROM golang

RUN mkdir -p $GOPATH/src/github.com/N-WhiteHole/tail/
ADD . $GOPATH/src/github.com/N-WhiteHole/tail/

# expecting to fetch dependencies successfully.
RUN go get -v github.com/N-WhiteHole/tail

# expecting to run the test successfully.
RUN go test -v github.com/N-WhiteHole/tail

# expecting to install successfully
RUN go install -v github.com/N-WhiteHole/tail
RUN go install -v github.com/N-WhiteHole/tail/cmd/gotail

RUN $GOPATH/bin/gotail -h || true

ENV PATH $GOPATH/bin:$PATH
CMD ["gotail"]
