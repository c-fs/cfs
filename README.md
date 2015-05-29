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

## Dependencies

cfs depends on GF-complete and Jerasure for erasure coding.

To install GF-complete, see https://github.com/c-fs/gf-complete

To install Jerasue, see https://github.com/c-fs/Jerasure/

TODO: write build/MAKE file

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
