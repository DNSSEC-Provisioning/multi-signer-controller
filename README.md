# multi-signer-controler
Control of a DNSSEC multi-signer group

## Notes

- Pick-up time for CDS/CDNSKEY on .ch may be up to one day / 24 hours

# local go

```
mkdir -p go/1.15.8; wget -O - https://storage.googleapis.com/golang/go1.15.8.linux-amd64.tar.gz | tar -C go/1.15.8 -zxv
export GOROOT="$HOME/go/1.15.8/go" GOPATH="$HOME/go"
export PATH="$PATH:$GOROOT"
```

# build

```
make
./multi-signer-controler
```

# Known Issues

- TSIG keys hardcoded to HMAC-SHA256
