# Keepass-HTTPD [![License](http://img.shields.io/:license-mit-blue.svg)](https://github.com/Gonzih/keepass-httpd/blob/master/LICENSE.md)

## Serve your keepass file via http api

### Not for production use yet

### Installation

```
go get github.com/Gonzih/keepass-httpd
```

### Usage

Start the server
```
$ keepass-httpd --keepass-file /path/to/file.kdbx --http-port 8080
```

Currently you always have to initialize db via http call
```
$ curl "http://localhost:8080/reload?password=mysecret"
> {"status":"success"}
```

Search endpoint uses logical AND for parameters to search
```
$ curl "http://localhost:8080/search?title=entrytitle&url=url.com&username=myusername"
> {"username":"myusername","title":"entrytitle","password":"securepassword","url":"url.com"}
```


Re-read DB file from the disk to memory
```
$ curl "http://localhost:8080/reload?password=mysecret"
> {"status":"success"}
```

To see all command line arguments run
```
keepass-httpd --help
```
