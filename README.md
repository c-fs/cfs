#cfs

## Play with cfs

### Run Server

``` bash
cd server

go build

./server

```

cfs listens on `15524` by default. The root path of cfs is the current directory by default

#### Write and Read file

``` bash
cd ctl

go build

./ctl -action write -name "foo" -data "bar"
2015/05/24 11:16:48 3 bytes written to foo at offset 0

./ctl -action read -name "foo" -length 100
2015/05/24 11:17:34 bar
```

#### Read a corrupted file

``` bash
echo "corrupt" -> ./server/foo
```

Try to read out the file again
``` 
./ctl/ctl -action read -name "foo" -length 100
2015/05/24 11:18:56 
```

Nothing is read out.

We can see the log output from the server side

```
2015/05/24 11:18:56 server: read error disk: not a valid CRC
```
