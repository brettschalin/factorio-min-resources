## Requirements

* A copy of Factorio version `1.1.77` or later - legal purchases are preferred for reasons I shouldn't need to explain
* The Go programming language - install from https://go.dev/dl/ or your package manager of choice (`1.18` or later is required)
* goyacc - run `go install golang.org/x/tools/cmd/goyacc@latest` and make sure `$(go env GOPATH)/bin` is in your `PATH`
* `make` and `patch` - these should come by default on Linux/Mac, but isn't hard to find either if you don't have them

## Get the data

As of version `1.1.77`, there's a command line option to dump the raw data the game uses as JSON - The TAS code should in theory work for much earlier versions but the Go command needs the data dump and therefore need this version or higher. Run `$FACTORIO_INSTALL_PATH/bin/x64/factorio --data-dump` and it'll dump a rather large (~35-40MB) JSON file into the `script-output` directory. Copy it to [`data`](./data).

## Map exchange string

This map has a good clustering of the starting resources that's reasonably close to water. If you find a better map please let me know

```
>>>eNpjZGBk8GVgYgCCBnsQ5mBJzk/MgfAOOIAwV3J+QUFqkW5+U
SqyMGdyUWlKqm5+Jqri1LzU3ErdpMRiqGKIyRyZRfl56CawFpfk5
6GKlBSlphbDnALC3KVFiXmZpbkIvVCnMi59/z+ioUWOAYT/1zMo/
P8PwkDWA6ACEGZgbICoBIrBAGtyTmZaGgODgiMQO4EVMTBWi6xzf
1g1xZ4RokbPAcr4ABU5kAQT8YQx/BxwSqnAGCZI5hiDwWckBsTSE
qAVUFUcDggGRLIFJMnI2Pt264Lvxy7YMf5Z+fGSb1KCPaOhq8i7D
0br7ICS7CAvMMGJWTNBYCfMKwwwMx/YQ6Vu2jOePQMCb+wZWUE6R
ECEgwWQOODNzMAowAdkLegBEgoyDDCn2cGMEXFgTAODbzCfPIYxL
tuj+wMYEDYgw+VAxAkQAbYQ7jJGCNOh34HRQR4mK4lQAtRvxIDsh
hSED0/CrD2MZD+aQzAjAtkfaCIqDliigQtkYQqceMEMdw0wPC+ww
3gO8x0YmUEMkKovQDEIDyQDMwpCCziAg5uZAQGAaUP20+XvAL9/o
5o=<<<
```

