package reverseproxy

import (
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
			pattern := concatPath(s.BaseUri, subPath)
			defaultProxyMux.handle(pattern, proxy, handler)

			// 如果 subPath == “/”，再配置一次做精确匹配
			if subPath == "/" {
				for _, method := range proxy.Methods {
					rb := newRouteBuilder(ws, method, subPath)
					ws.Route(rb.To(onMessage))
				}
			}

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
