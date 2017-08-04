# Keepass-HTTPD [![License](http://img.shields.io/:license-mit-blue.svg)](https://github.com/Gonzih/keepass-httpd/blob/master/LICENSE.md)

## Serve your keepass file via http api

### Not for production use yet

### Installation

```
go get github.com/Gonzih/keepass-httpd
```

### Usage:

```
$ keepass-httpd --keepass-file /path/to/file.kdbx --keepass-password mysecret
```

```
# Uses logical AND for parameters to search
$ curl "http://localhost:8080/search?title=entrytitle&url=url.com&username=myusername"
> {"username":"myusername","title":"entrytitle","password":"securepassword","url":"url.com"}


# Reloads DB file from the disk to memory
curl -X POST "http://localhost:8080/reload"
> {"status":"success"}
```

```
# To see all command line arguments
keepass-httpd --help
```
