package reverseproxy

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/emicklei/go-restful"
)

func NewWebService(dir string) (wss []*restful.WebService, err error) {
	files, err := walkYaml(dir)
	if err != nil {
		return
	}

	var s server
	for _, f := range files {
		if err = parseYaml(f, &s); err != nil {
			return
		}

		ws := new(restful.WebService)
		ws.Path(s.BaseUri)

		handler := newSingleHostReverseProxy(s.Target)
		for subPath, proxy := range s.Proxy {
			// 如果 subPath == “/”，再配置一次做精确匹配
			if subPath == "/" {
				pattern := strings.TrimRight(s.BaseUri, "/")
				defaultProxyMux.handle(pattern, proxy, handler)
				for _, method := range proxy.Methods {
					rb := newRouteBuilder(ws, method, subPath)
					ws.Route(rb.To(onMessage))
				}
			}

			pattern := concatPath(s.BaseUri, subPath)
			defaultProxyMux.handle(pattern, proxy, handler)

			if strings.HasSuffix(subPath, "/") {
				subPath += "{subpath:*}"
			}
			for _, method := range proxy.Methods {
				rb := newRouteBuilder(ws, method, subPath)
				ws.Route(rb.To(onMessage))
			}
		}
		wss = append(wss, ws)
		s.reset()
	}
	return
}

func onMessage(req *restful.Request, resp *restful.Response) {
	fmt.Printf("req.Request.URL.Path = %v\n", req.Request.URL.Path)
	fmt.Printf("req.Request.URL.RawPath = %v\n", req.Request.URL.RawPath)
	fmt.Printf("req.Request.URL.RawQuery = %v\n", req.Request.URL.RawQuery)

	fmt.Printf("req.pathParameters = %v\n", req.PathParameters())
	fmt.Printf("req.selectedRoutePath = %v\n", req.SelectedRoutePath())
	subPath := req.PathParameter("subpath")
	fmt.Printf("subpath: %s\n", subPath)
	pattern, proxy, handler := defaultProxyMux.match(req.Request.URL.Path)
	if len(pattern) == 0 {
		result := make(map[string]string)
		result["message"] = "not found"
		result["code"] = "500"
		resp.WriteHeaderAndEntity(404, result)
	}
	fmt.Println("proxy mux match", pattern, proxy.Pass)

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
