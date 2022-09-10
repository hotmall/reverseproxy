package reverseproxy

import (
	"net/url"
	"strings"

	"github.com/emicklei/go-restful"
)

func NewWebService(dir string) (ws *restful.WebService, err error) {
	files, err := walkYaml(dir)
	if err != nil {
		return
	}

	ws = new(restful.WebService)
	var s server
	for _, f := range files {
		if err = parseYaml(f, &s); err != nil {
			return
		}

		handler := newSingleHostReverseProxy(s.Target)
		for pattern, proxy := range s.Proxy {
			// fmt.Printf("pattern: %s\n", pattern)
			pattern := concatPath(s.BaseUri, pattern)
			defaultProxyMux.handle(pattern, proxy, handler)
			for _, method := range proxy.Methods {
				if strings.HasSuffix(pattern, "/") {
					pattern += "{subpath:*}"
				}
				rb := newRouteBuilder(ws, method, pattern)
				ws.Route(rb.To(onMessage))
			}
		}
		s.reset()
	}
	return
}

func onMessage(req *restful.Request, resp *restful.Response) {
	// fmt.Println("on message", "xx", req.Request.URL.Path, req.Request.URL.RawPath, req.Request.URL.RawQuery)
	// fmt.Println("routePath", req.SelectedRoutePath())
	subPath := req.PathParameter("subpath")
	// fmt.Printf("subpath: %s\n", subPath)
	_, proxy, handler := defaultProxyMux.match(req.Request.URL.Path)
	// fmt.Println("proxy mux match", pattern, proxy.Pass)

	a, err := url.Parse(proxy.Pass)
	if err != nil {
		result := make(map[string]string)
		result["message"] = err.Error()
		result["code"] = "500"
		resp.WriteHeaderAndEntity(500, result)
		return
	}
	b, err := url.Parse(subPath)
	if err != nil {
		result := make(map[string]string)
		result["message"] = err.Error()
		result["code"] = "500"
		resp.WriteHeaderAndEntity(500, result)
		return
	}
	req.Request.URL.Path, req.Request.URL.RawPath = joinURLPath(a, b)
	handler.ServeHTTP(resp.ResponseWriter, req.Request)
}

func concatPath(path1, path2 string) string {
	return strings.TrimRight(path1, "/") + "/" + strings.TrimLeft(path2, "/")
}

func newRouteBuilder(ws *restful.WebService, method, subPath string) *restful.RouteBuilder {
	return ws.Method(method).Path(subPath)
}
