SOURCES := $(wildcard *.go)

all: multi-signer-controller

fmt: format

format:
	gofmt -w *.go
	sed -i -e 's%	%    %g' *.go

multi-signer-controller: $(SOURCES)
	go build -v -x
