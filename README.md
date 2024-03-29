# Reverse Proxy

## install

```shell
go get github.com/hotmall/reverseproxy
```

## Usage

```go
import "github.com/emicklei/go-restful"

wss, err := reverseproxy.NewWebService("./etc/proxy")
if err != nil {
    panic("new revese proxy webservice fail")
}
for _, ws := range wss {
  restful.Add(ws)
}
```

YAML file configuration

```yaml
title: Reverse Proxy Configuration
version: v1
baseUri: /ops/{version}/dynconf
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

Note:

```yaml
proxy:
  /global/:
    methods:
      - GET
      - PUT
    pass:
      /dynconf/{version}/global
```

Subpath `/global/` means url pattern, must begin with `/`, and end with `/` means prefix match, no `/` suffix means exact match.

```yaml
proxy:
  /:
    methods:
      - GET
      - PUT
    pass:
      /pay/v1/banks
```

Subpath `/` means prefix match and exact match.
