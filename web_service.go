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
		ws.Path(s.BaseUri).Consumes("application/json").Produces("application/json")

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
	// subPath := req.PathParameter("subpath")
	// fmt.Printf("subpath: %s\n", subPath)
	// routePath =
	pattern, proxy, handler := defaultProxyMux.match(req.SelectedRoutePath())
	if len(pattern) == 0 {
		e := acquireError()
		defer releaseError(e)
		e.Code = 404
		e.Message = "proxy: not found route"
		resp.WriteHeaderAndEntity(404, e)
		return
	}
	fmt.Println("proxy mux match", pattern, proxy.Pass)

	// 判断是否存在路径参数，存在路径参数，构建 pass
	params := req.PathParameters()
	if v, ok := params["subpath"]; ok {
		subPath := v
		a, err := url.Parse(proxy.Pass)
		if err != nil {
			e := acquireError()
			defer releaseError(e)
			e.Code = 500
			e.Message = err.Error()
			resp.WriteHeaderAndEntity(500, e)
			return
		}
		b, err := url.Parse(subPath)
		if err != nil {
			e := acquireError()
			defer releaseError(e)
			e.Code = 500
			e.Message = err.Error()
			resp.WriteHeaderAndEntity(500, e)
			return
		}
		req.Request.URL.Path, req.Request.URL.RawPath = joinURLPath(a, b)
	} else {
		// 判断 proxy_pass 是否含有路径参数
		pos := strings.Index(proxy.Pass, "{")
		if pos != -1 {
			// 含有路径参数替换之
			tokens := tokenizePath(proxy.Pass)
			for ind, each := range tokens {
				if strings.HasPrefix(each, "{") {
					varName := strings.TrimSpace(each[1 : len(each)-1])
					if v, ok := params[varName]; ok {
						tokens[ind] = v
					}
				}
			}
			req.Request.URL.Path = strings.Join(tokens, "/")
		} else {
			req.Request.URL.Path = proxy.Pass
		}
	}

	fmt.Printf("req.Request.URL.Path222 = %v\n", req.Request.URL.Path)

	handler.ServeHTTP(resp.ResponseWriter, req.Request)
}

func concatPath(path1, path2 string) string {
	return strings.TrimRight(path1, "/") + "/" + strings.TrimLeft(path2, "/")
}

func tokenizePath(path string) []string {
	if path == "/" {
		return nil
	}
	return strings.Split(strings.Trim(path, "/"), "/")
}

func newRouteBuilder(ws *restful.WebService, method, subPath string) *restful.RouteBuilder {
	return ws.Method(method).Path(subPath)
}
