#Cloud File System
[![Build Status](https://travis-ci.org/c-fs/cfs.png?branch=master)](https://travis-ci.org/c-fs/cfs)

## Overview
cfs (cloud file system) is a low-level filesystem interface for building the cloud. cfs is not a distributed file system. It is designed to be the lowest building block of cloud infrastructure that is available on all machines. The design goal of cfs is reliability, simplicity and usability.

## Status

Current status is pre-alpha

TODO:

1. basic file system interface (write, read, remove, mkdir, rename)
2. copy from another cfs server
3. bit-matrix based data reconstruction
4. expose basic run time metrics
5. basic resource enforcement 
6. container deployment

alpha

TODO:

1. profiling and benchmark
2. per disk metadata store
3. simple ACL support
4. user level buffer
5. tons of tests

We will **BREAK** everything (data format, API, command line tool, etc.) before 1.0.

## Install

cfs depends on [GF-complete](https://github.com/c-fs/gf-complete) and [Jerasure](https://github.com/c-fs/Jerasure/) for erasure coding.

cfs requires that Go version must be 1.5 at minimum.

For now, to install these packages:
``` bash
go get github.com/c-fs/vendor
go get -u github.com/c-fs/Jerasure
export GO15VENDOREXPERIMENT=1
make build
```

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

#### Run Server in a Docker Container

You can run a cfs to provide file service in container easily:

```
make docker

sudo docker run \
  --volume=/:/rootfs:ro \
  --volume=/var/run:/var/run:rw \
  --volume=/sys:/sys:ro \
  --volume=/var/lib/docker/:/var/lib/docker:ro \
  --volume=/tmp:/tmp:rw \
  --volume=${PWD}/cfs0000:/cfs0000
  --volume=${PWD}/cfs0001:/cfs0001
  --publish=15524:15524 \
  --detach=true \
  --name=cfs \
  c-fs/cfs
```

cfs is now running in the background on `http://localhost:15524`, and it uses ./cfs0000/ and ./cfs0001/ as data volume. For development, you can replace `--detach=true` with `-ti` to let it run as an interactive process.

#### Write and Read file

``` bash
cd cfsctl

go build

cfsctl write --name="cfs0/foo" --data="bar"
2015/05/24 11:16:48 3 bytes written to foo at offset 0

cfsctl read --name="cfs0/foo" --length=100
2015/05/24 11:17:34 bar
```

#### Rename and Remove file

``` bash
cfsctl rename --oldname="cfs0/foo" --newname="cfs0/food"
2015/05/28 15:20:22 rename cfs0/foo into cfs0/food

cfsctl remove --name="cfs0/food"
2015/05/28 15:20:43 deletion succeeded
```

#### Read a corrupted file

``` bash
echo "corrupt" -> ./server/cfs0000/foo
```

Try to read out the file again
```
cfsctl read --name="cfs0/foo" --length=100
2015/05/24 11:18:56
```

Nothing is read out.

We can see the log output from the server side

```
2015/05/24 11:18:56 server: read error disk: not a valid CRC
```
