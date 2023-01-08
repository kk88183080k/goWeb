package msgo

import (
	"errors"
	"fmt"
	"github.com/kk88183080k/goWeb/msgo/logs"
	"github.com/kk88183080k/goWeb/msgo/msconf"
	"github.com/kk88183080k/goWeb/msgo/render"
	"github.com/kk88183080k/goWeb/msgo/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const ANY = "ANY"

// 中间件定义 start

type Handler func(ctx *Context)
type MiddlewareFun func(handler Handler) Handler
type ErrorHandlerFun func(err error) (int, any) // 返回http状态码，及返回前端的数据

// 中间件定义 start

type routerGroup struct {
	Name          string
	PathMap       map[string]map[string]Handler         // k=/ , v= {k:method, v:Handler=处理函数}
	MiddlePathMap map[string]map[string][]MiddlewareFun // k=/ , v= {k:method, v: []MiddlewareFun 函数前后执行的中间件}
	PathMethodMap map[string][]string                   // k = any|get|post  v:= []{"/1"， /2}
	treeNode      *TreeNode                             // 前缀树
	middlewares   []MiddlewareFun                       // 中间件
}

func (rg *routerGroup) Use(middles ...MiddlewareFun) {
	// 添加分组级别的中间件
	rg.middlewares = append(rg.middlewares, middles...)
}

func (r *routerGroup) methodHandle(name string, method string, handler Handler, ctx *Context) {
	// 分组级的中间件执行
	if r.middlewares != nil {
		for _, midd := range r.middlewares {
			handler = midd(handler)
		}
	}

	// 方法级的中间件执行
	midds := r.MiddlePathMap[name][method]
	if midds != nil {
		for _, m := range midds {
			handler = m(handler)
		}
	}
	// 最终执行的业务方法
	handler(ctx)
}

// groupName 必须以/开头
func (r *router) Group(groupName string) *routerGroup {
	// 先从已有的分组中查找
	for _, routerGroup := range r.RouterGroup {
		if routerGroup.Name == groupName {
			return routerGroup
		}
	}
	// 没有时再生成一个新的
	tree := &TreeNode{}
	tree.name = "/"
	group := &routerGroup{
		Name:          groupName,
		PathMap:       make(map[string]map[string]Handler, 0),
		MiddlePathMap: make(map[string]map[string][]MiddlewareFun, 0),
		PathMethodMap: make(map[string][]string, 0),
		treeNode:      tree,
		middlewares:   make([]MiddlewareFun, 0)}
	r.RouterGroup = append(r.RouterGroup, group)
	// 设置全局中间件
	group.middlewares = append(group.middlewares, r.middlewares...)
	return group
}

// 带*号的路由地址得放在尾部，否则会影响精准匹配
func (rg *routerGroup) add(method, api string, handlerFn Handler, midFn ...MiddlewareFun) *routerGroup {
	// 添加函数map
	_, ok := rg.PathMap[api]
	if !ok {
		rg.PathMap[api] = make(map[string]Handler, 0)
		rg.MiddlePathMap[api] = make(map[string][]MiddlewareFun)
	}
	_, ok = rg.PathMap[api][method]
	if ok {
		panic("有重复的路由")
	}

	// 设置请求处理的函数
	rg.PathMap[api][method] = handlerFn
	// 添加请求方式
	apiMethodArray, ok := rg.PathMethodMap[method]
	for _, v := range apiMethodArray {
		if v == api {
			panic(api + " method, is exists")
		}
	}
	rg.PathMethodMap[method] = append(apiMethodArray, api)

	// 添加方法级的中间件
	if midFn != nil {
		rg.MiddlePathMap[api][method] = append(rg.MiddlePathMap[api][method], midFn...)
	}

	nodeUrl := rg.Name + api
	rg.treeNode.Put(nodeUrl)
	//log.Println("tree put :", rg.Name, nodeUrl, api)

	return rg
}

func (rg *routerGroup) Any(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(ANY, api, handler, midFn...)
}

func (rg *routerGroup) Post(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(http.MethodPost, api, handler, midFn...)
}

func (rg *routerGroup) Get(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(http.MethodGet, api, handler, midFn...)
}

func (rg *routerGroup) Put(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(http.MethodPut, api, handler, midFn...)
}

func (rg *routerGroup) Patch(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(http.MethodPatch, api, handler, midFn...)
}

func (rg *routerGroup) Options(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(http.MethodOptions, api, handler, midFn...)
}

func (rg *routerGroup) Head(api string, handler Handler, midFn ...MiddlewareFun) *routerGroup {
	return rg.add(http.MethodHead, api, handler, midFn...)
}

type router struct {
	RouterGroup []*routerGroup
	middlewares []MiddlewareFun // 中间件
}

type Engine struct {
	*router
	fnMap      template.FuncMap
	render     render.HTMLRender
	pool       sync.Pool
	logger     *logs.Logger
	errHandler ErrorHandlerFun
}

func New() *Engine {
	r := &router{RouterGroup: []*routerGroup{}}
	e := &Engine{router: r, fnMap: template.FuncMap{}, logger: logs.Default()}
	e.pool.New = func() any {
		log.Println("create Context success")
		return &Context{e: e}
	}
	e.errHandler = func(err error) (int, any) {
		switch er := err.(type) {
		case *R:
			return http.StatusOK, er.Response()
		default:
			return http.StatusInternalServerError, "Internal Server Error"
		}
	}

	// 根据配置设置日志文件所在的目录
	logPath, ok := msconf.Conf.Log["path"]
	if ok {
		e.logger.SetPath(logPath.(string))
	}

	return e
}

func Default() *Engine {
	// 设置默认的中间件
	return New().Use(Recovery, Logging)
}

func (e *Engine) SetFnMap(fnMap template.FuncMap) {
	e.fnMap = fnMap
}

func (e *Engine) SetRender(t *template.Template) {
	e.render = render.HTMLRender{Template: t}
}

func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.fnMap).ParseGlob(pattern))
	e.SetRender(t)
}

func (e *Engine) LoadTemplateByConf() {
	confPattern, ok := msconf.Conf.Template["pattern"]
	if !ok {
		panic(errors.New("Template pattern is not config "))
	}
	t := template.Must(template.New("").Funcs(e.fnMap).ParseGlob(confPattern.(string)))
	e.SetRender(t)
}

func (e *Engine) RegisterErrorHandler(handler ErrorHandlerFun) {
	e.errHandler = handler
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := e.pool.Get().(*Context)
	// 设置初始值，否则会缓存
	context.W = w
	context.R = r
	context.queryCache = nil
	context.formCache = nil
	context.DisallowUnknownFields = false
	context.IsValidate = false
	context.StatusCode = -1
	context.Logger = e.logger

	e.severHttpRequestHandle(context)

	e.pool.Put(context)
}

func (e *Engine) severHttpRequestHandle(ctx *Context) {
	r := ctx.R
	w := ctx.W
	method := r.Method
	path := r.URL.Path
	for _, rg := range e.RouterGroup {
		node := rg.treeNode.Get(path)
		//log.Printf("tree get, rg.Name:%s, urlPath: %s, treeNode: %v\n", rg.Name, path, node)

		if node != nil && node.leaf {
			apiUrl := utils.SubStringLast(node.routerFullPath, rg.Name)
			//log.Printf("method match: %v\n", apiUrl)
			// 先匹配any的
			handle, ok := rg.PathMap[apiUrl][ANY]
			if ok {
				//handle(ctx)
				rg.methodHandle(apiUrl, ANY, handle, ctx)
				return
			}
			// 再匹配其他的
			handle, ok = rg.PathMap[apiUrl][method]
			if ok {
				//handle(ctx)
				rg.methodHandle(apiUrl, ANY, handle, ctx)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s %s not allowed \n", r.RequestURI, method)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%v not find", http.StatusNotFound)
}

func (e *Engine) Start(ip string, port int) {
	localAddress := ip + ":" + strconv.Itoa(port)
	log.Printf("start server by : http://%s\n", localAddress)
	err := http.ListenAndServe(localAddress, e)
	if err != nil {
		panic(err)
	}
}

// StartByTLS https 支持
func (e *Engine) StartByTLS(addr, certFile, keyFile string) {
	log.Printf("start server by : https://%s\n", addr)
	err := http.ListenAndServeTLS(addr, certFile, keyFile, e)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) Use(fn ...MiddlewareFun) *Engine {
	e.router.middlewares = append(e.router.middlewares, fn...)
	return e
}
