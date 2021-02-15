SOURCES := $(wildcard *.go)

all: multi-signer-controler

fmt: format

format:
	gofmt -w *.go
	sed -i -e 's%	%    %g' *.go

multi-signer-controler: $(SOURCES)
	go build -v -x
