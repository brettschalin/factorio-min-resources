## Requirements

* A copy of Factorio version `1.1.77` or later - legal purchases are preferred for reasons I shouldn't need to explain
* The Go programming language - install from https://go.dev/dl/ or your package manager of choice (`1.18` or later is required)

## Get the data

As of version `1.1.77`, there's a command line option to dump the raw data the game uses as JSON - The TAS code should in theory work for much earlier versions but the Go command needs the data dump and therefore need this version or higher. Run `$FACTORIO_INSTALL_PATH/bin/x64/factorio --data-dump` and it'll dump a rather large (~35-40MB) JSON file into the `script-output` directory. Copy it to [`data`](./data).

## Map exchange string

This map has a good clustering of the starting resources and has oil reasonably close to water. If you find a better map please let me know

```
>>>eNpjZGBkiABiIGiwB2EOluT8xBwGhgMOMMyVnF9QkFqkm1+Ui
izMmVxUmpKqm5+Jqjg1LzW3UjcpsTgVYiLEZI7Movw8dBNYi0vy8
1BFSopSU4uRNXKXFiXmZZbmQvQixBkYl77/H9HQIscAwv/rGRT+/
wdhIOsBUAEIMzA2QFQCxaCASTY5P6+kKD9Htzi1pCQzL90qNz+zu
KS0KNUqKTOxmMNAz9QABHRxKksrSi0sTc1LrrTKLc0pySzIyUwtA
mozNAMCc9bknMy0NAYGBUcgdgI7gYGxWmSd+8OqKfaMECfoOUAZH
6AiB5JgIp4whp8DTikVGMMEyRxjMPiMxIBYWgK0AqqKwwHBgEi2g
CQZGXvfbl3w/dgFO8Y/Kz9e8k1KsGc0dBV598FonR1Qkh3kBSY4M
WsmCOyEeYUBZuYDe6jUTXvGs2dA4I09IytIhwiIcLAAEge8mRkYB
fiArAU9QEJBhgHmNDuYMSIOjGlg8A3mk8cwxmV7dH8AA8IGZLgci
DgBIsAWwl3GCGE69DswOsjDZCURSoD6jRiQ3ZCC8OFJmLWHkexHc
whmRCD7A01ExQFLNHCBLEyBEy+Y4a4BhucFdhjPYb4DIzOIAVL1B
SgG4YFkYEZBaAEHcHDDZKFpI/PitU4AtZ/FiQ==<<<
```
