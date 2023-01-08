package msgo

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/kk88183080k/goWeb/msgo/binding"
	"github.com/kk88183080k/goWeb/msgo/logs"
	"github.com/kk88183080k/goWeb/msgo/render"
	"github.com/kk88183080k/goWeb/msgo/utils"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	e                     *Engine
	queryCache            url.Values   // get请求，地址中的参数
	formCache             url.Values   // post请求，body中的参数
	DisallowUnknownFields bool         // 客户端传的参数中有，但后台结构体中没有就报错
	IsValidate            bool         // 客户端传的参数是否校验
	StatusCode            int          // 返回的状态码
	Logger                *logs.Logger // 日志组件
}

// 解析表单使用的最大内存参数
const defaultMultipartMemory = 2 << 16

/*****原始写法** strart ***/

func (c *Context) Html(status int, html string) error {
	c.W.WriteHeader(status)
	c.W.Header().Add("Content-type", "text/html; charset=utf-8")
	_, err := c.W.Write([]byte(html))
	return err
}

// 按文件名解析
func (c *Context) HtmlTemplateNoLoad(name string, funcMap template.FuncMap, status int, data any, tFileName ...string) error {
	c.W.Header().Add("Content-type", "text/html; charset=utf-8")
	c.W.WriteHeader(status)

	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseFiles(tFileName...)
	if err != nil {
		return err
	}

	return t.Execute(c.W, data)
}

// 按正则表达式匹配
func (c *Context) HtmlTemplateGlobNoLoad(name string, funcMap template.FuncMap, status int, data any, pattern string) error {
	c.W.Header().Add("Content-type", "text/html; charset=utf-8")
	c.W.WriteHeader(status)

	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}

	return t.Execute(c.W, data)
}

// 按文件名解析
func (c *Context) HtmlTemplate(status int, name string, data any) error {
	c.W.Header().Add("Content-type", "text/html; charset=utf-8")
	c.W.WriteHeader(status)
	return c.e.render.Template.ExecuteTemplate(c.W, name, data)
}

func (c *Context) JSON(staus int, data any) error {
	c.W.Header().Add("Content-Type", "application/json; charset=utf-8")
	c.W.WriteHeader(staus)

	dataByte, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(dataByte)
	return err
}

func (c *Context) XML(status int, data any) error {
	c.W.Header().Add("Content-Type", "application/xml; charset=utf-8")
	c.W.WriteHeader(status)
	return xml.NewEncoder(c.W).Encode(data)
}

// File 文件下载
func (c *Context) File(filepath string) {
	http.ServeFile(c.W, c.R, filepath)
}

// FileAttachment 指定下载的文件名
func (c *Context) FileAttachment(filePath string, downloadFileName string) {
	if utils.IsASCII(downloadFileName) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+downloadFileName+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(downloadFileName))
	}
	http.ServeFile(c.W, c.R, filePath)
}

func (c *Context) FileFormFs(filePath string, fs http.FileSystem) {

	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filePath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

func (c *Context) Redirect(status int, url string) {
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		log.Println("Redirect status 异常")
		return
	}
	http.Redirect(c.W, c.R, url, status)
}

func (c *Context) String(status int, format string, data ...any) error {
	c.W.Header().Add("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(status)
	if len(data) > 0 {
		_, err := fmt.Fprintf(c.W, format, data...)
		return err
	}

	_, err := c.W.Write(utils.StringToBytes(format))
	return err
}

func (c *Context) Fail(status int, format string) error {
	return c.String(status, format)
}

/*****原始写法** end ***/

/*****接口抽象写法** start ***/

// Render 公共的解析方法
func (c *Context) Render(statusCode int, w http.ResponseWriter, viewResv render.Render) error {
	// 视图解析器中，设置content-type, 返回数据
	c.StatusCode = statusCode
	return viewResv.Render(w)
}

func (c *Context) StringOptions(status int, format string, data ...any) error {
	return c.Render(status, c.W, &render.String{Format: format, Data: data})
}

func (c *Context) XmlOptions(status int, data any) error {
	return c.Render(status, c.W, &render.Xml{Data: data})
}

func (c *Context) JsonOptions(status int, data any) error {
	return c.Render(status, c.W, &render.Json{Data: data})
}

func (c *Context) RedirectOptions(status int, url string) error {
	return c.Render(status, c.W, &render.Redirect{Url: url, Status: status, Request: c.R})
}

func (c *Context) HtmlOptions(status int, data string) error {
	c.W.WriteHeader(status)
	return c.Render(status, c.W, &render.HtmlOptionsRender{Name: "", Data: data, Template: c.e.render.Template, IsTemplate: false})
}

func (c *Context) HtmlTemplateOptions(status int, name string, data any) error {
	c.W.WriteHeader(status)
	return c.Render(status, c.W, &render.HtmlOptionsRender{Name: name, Data: data, Template: c.e.render.Template, IsTemplate: true})
}

/*****接口抽象写法** end ***/

/*****get 方式获取请求参数** start ***/
func (c *Context) initQueryCache() {
	if c.queryCache == nil {
		if c.R != nil {
			c.queryCache = c.R.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}
}

func (c *Context) DefaultQuery(key, defaultVal string) string {
	c.initQueryCache()
	val, ok := c.queryCache[key]
	if !ok {
		return defaultVal
	}
	return val[0]
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.DefaultQuery(key, "")
}

func (c *Context) GetArray(key string) []string {
	c.initQueryCache()
	return c.queryCache[key]
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) GetQueryMap(key string) map[string]string {
	dic, _ := c.GetMap(key)
	return dic
}

func (c *Context) GetMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

// get
func (c *Context) get(allMap map[string][]string, key string) (map[string]string, bool) {
	c.initQueryCache()
	//user[id]=1&user[name]=张三

	dicMap := make(map[string]string, 0)
	exists := false
	for k, v := range allMap {
		// 第一个[的位置
		i := strings.IndexByte(k, '[')
		if i >= 1 && k[0:i] == key {
			j := strings.IndexByte(k[i+1:], ']')
			if j >= 1 {
				exists = true
				mapkey := k[i+1:][:j]
				dicMap[mapkey] = v[0]
			}
		}
	}

	return dicMap, exists
}

/*****get 方式获取请求参数** end ***/

/*****post 方式获取请求参数** start ***/
func (c *Context) initFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil && errors.Is(err, http.ErrNotMultipart) {
			log.Println("解析表单出错", err)
			return
		}
		c.formCache = c.R.PostForm
	}

}

func (c *Context) GetForm(key string) string {
	c.initFormCache()
	return c.formCache.Get(key)
}

func (c *Context) GetFormArray(key string) (rs []string, ok bool) {
	c.initFormCache()
	rs, ok = c.formCache[key]
	return
}

func (c *Context) GetFormArrayVal(key string) (rs []string) {
	c.initFormCache()
	rs, _ = c.formCache[key]
	return
}

func (c *Context) GetFormMap(key string) (formMap map[string]string, ok bool) {
	c.initFormCache()
	formMap, ok = c.get(c.formCache, key)
	return
}

/*****post 方式获取请求参数** end ***/

/*****post 文件上传方式获取请求参数** start ***/
func (c *Context) GetFormFile(fileKey string) (*multipart.FileHeader, error) {
	if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
		return nil, err
	}

	file, header, err := c.R.FormFile(fileKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("close file error ", err)
		}
	}()

	return header, nil
}

func (c *Context) UploadFile(fileKey, dir string) (string, error) {
	file, err := c.GetFormFile("file")
	if err != nil {
		return "", err
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	extName := ""
	if i := strings.IndexByte(file.Filename, '.'); i >= 1 {
		extName = file.Filename[i:]
	}
	fileName := dir + "/" + time.Now().Format("20060102150405") + extName
	des, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer des.Close()

	buf := make([]byte, 1024)
	_, err = io.CopyBuffer(des, src, buf)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

/*****post 文件上传方式获取请求参数** end ***/

/*****post json方式获取请求参数** start ***/

func (c *Context) BindJson(obj any) error {
	jsonBinding := binding.JsonBind
	jsonBinding.DisallowUnknownFields = c.DisallowUnknownFields
	jsonBinding.IsValidate = c.IsValidate
	return c.MustBindWith(obj, &jsonBinding)
}

func (c *Context) BindXml(obj any) error {
	return c.MustBindWith(obj, &binding.XmlBind)
}

func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	//如果发生错误，返回400状态码 参数错误
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}

	return nil
}

func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.R, obj)
}

/*****post json方式获取请求参数** end ***/

/*****错误处理** start ***/

func (c *Context) ErrorHandler(err error) {
	c.JsonOptions(c.e.errHandler(err))
}

func (c *Context) HandlerWithError(code int, msg string, err error) {
	if err != nil {
		c.JsonOptions(c.e.errHandler(err))
		return
	}

	c.JsonOptions(code, msg)
}

/*****错误处理** end ***/
