## Overview
cfs (cloud file system) is a low-level filesystem interface for building the cloud. cfs is not a distributed file system. It is designed to be the lowest building block of cloud infrastructure that is available on all machines. The design goal of cfs is reliability, simplicity and usability.

## Status

pre-alpha

We will **BREAK** everything (data format, API, command line tool, etc.) before 1.0.

## Play with cfs

### Run Server

``` bash
cd server

go build

./server

```

cfs listens on `15524` by default. The root path of cfs is the current directory by default.
cfs creates a `/cfs0` disk with root path `/cfs0000` by default. 
TODO: make all these configurable.

#### Write and Read file

``` bash
cd ctl

go build

./ctl -action write -name "cfs0/foo" -data "bar"
2015/05/24 11:16:48 3 bytes written to foo at offset 0

./ctl -action read -name "cfs0/foo" -length 100
2015/05/24 11:17:34 bar
```

#### Read a corrupted file

``` bash
echo "corrupt" -> ./server/cfs0000/foo
```

Try to read out the file again
``` 
./ctl/ctl -action read -name "cfs0/foo" -length 100
2015/05/24 11:18:56 
```

Nothing is read out.

We can see the log output from the server side

```
2015/05/24 11:18:56 server: read error disk: not a valid CRC
```
