# Reverse Proxy

## install

```shell
go get github.com/hotmall/reverseproxy
```

## Usage

```go
import "github.com/emicklei/go-restful"

ws, err := reverseproxy.NewWebService("./etc/proxy")
if err != nil {
    panic("new revese proxy webservice fail")
}
restful.Add(ws)

```

YAML file configuration

```yaml
title: Reverse Proxy Configuration
version: v1
baseUri: /ops/{version}
target: http://127.0.0.1:8000
proxy:
  /global/:
    methods:
      - GET
      - PUT
    pass:
      /dynconf/{version}/global

  /services/:
    methods:
      - GET
      - PUT
    pass:
      /dynconf/{version}/services

  /nodes/:
    methods:
      - GET
      - PUT
    pass:
      /dynconf/{version}/nodes
```
