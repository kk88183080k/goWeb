package main

import (
	"errors"
	"fmt"
	"github.com/kk88183080k/goWeb/msgo"
	"github.com/kk88183080k/goWeb/msgo/logs"
	"github.com/kk88183080k/goWeb/msgo/mserror"
	"github.com/kk88183080k/goWeb/msgo/mspool"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type user struct {
	Name    string   `xml:"name" json:"name" msgo:"required"`
	Age     int      `xml:"age,string" json:"age,int"  msgo:"required" validate:"required,max=50,min=18"`
	Address []string `xml:"address" json:"address"  msgo:"required"`
	Email   string   `xml:"email"json:"email"`
}

func main() {

	engine := msgo.Default()
	// 注册全局异常处理函数
	//engine.RegisterErrorHandler(func(err error) (int, any) {
	//
	//})

	//reqResponseTest(engine)
	//errorTest(engine)
	goPoolTest(engine)

	//engine.LoadTemplate("tpl/*.html")
	engine.LoadTemplateByConf()
	engine.Start("127.0.0.1", 8080)
	//engine.StartByTLS("127.0.0.1:8888", "cert/server.pem", "cert/server.key")
}

func goPoolTest(engine *msgo.Engine) {

	group := engine.Group("/pool")
	group.Get("/testGo", func(ctx *msgo.Context) {
		currentTime := time.Now().UnixMilli()
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)

		go func() {
			time.Sleep(time.Duration(3) * time.Second)
			ctx.Logger.Info("执行函数1")
			waitGroup.Done()
		}()
		go func() {
			time.Sleep(time.Duration(3) * time.Second)
			ctx.Logger.Info("执行函数2")
			waitGroup.Done()
		}()
		waitGroup.Wait()
		fmt.Printf("time: %v \n", time.Now().UnixMilli()-currentTime)
		ctx.JsonOptions(http.StatusOK, "ok")
	})

	pool, _ := mspool.NewDefaultPool()
	group.Get("/test", func(ctx *msgo.Context) {
		currentTime := time.Now().UnixMilli()
		var waitGroup sync.WaitGroup
		waitGroup.Add(5)
		sleepTime := 2
		pool.Submit(func() {
			ctx.Logger.Info("执行函数1")
			time.Sleep(time.Duration(sleepTime) * time.Second)
			waitGroup.Done()
		})
		pool.Submit(func() {
			ctx.Logger.Info("执行函数2")
			time.Sleep(time.Duration(sleepTime) * time.Second)
			waitGroup.Done()
		})

		pool.Submit(func() {
			ctx.Logger.Info("执行函数3")
			time.Sleep(time.Duration(sleepTime) * time.Second)
			waitGroup.Done()
		})

		pool.Submit(func() {
			ctx.Logger.Info("执行函数4")
			time.Sleep(time.Duration(sleepTime) * time.Second)
			waitGroup.Done()
		})

		pool.Submit(func() {
			ctx.Logger.Info("执行函数5")
			time.Sleep(time.Duration(sleepTime) * time.Second)
			waitGroup.Done()
		})

		waitGroup.Wait()
		fmt.Printf("time: %v \n", time.Now().UnixMilli()-currentTime)
		ctx.JsonOptions(http.StatusOK, "ok")
	})
}

// errorTest 验证错误
func errorTest(engine *msgo.Engine) {
	group := engine.Group("/recover")

	// 原生异常
	b := 1
	group.Get("/test", func(ctx *msgo.Context) {
		if b == 1 {
			panic(errors.New("异常了，我要退出"))
		}
		fmt.Println("异常了，我要退出")
	})

	// 自定义错误
	group.Get("/customErr", func(ctx *msgo.Context) {
		msError := mserror.Default()
		msError.Result(func(e *mserror.MsError) {
			ctx.Logger.Error(e.Error())
			ctx.JSON(http.StatusInternalServerError, "服务器内部错误")
		})
		msError.Put(errors.New("我异常了"))
	})

	// 自定义返回结果
	group.Get("/response", func(ctx *msgo.Context) {
		ctx.ErrorHandler(errors.New("自定义返回结果第一个string错误"))
	})
	group.Get("/response1", func(ctx *msgo.Context) {
		r := msgo.DefaultR()
		r.Code = 200
		r.Msg = "操作成功"
		r.Data = "我是数据"
		ctx.ErrorHandler(r)
	})
	group.Get("/response2", func(ctx *msgo.Context) {
		r := msgo.DefaultR()
		r.Code = 500
		r.Msg = "操作失败"
		ctx.ErrorHandler(r)
	})

	group.Get("/response3", func(ctx *msgo.Context) {
		ctx.ErrorHandler(msgo.DefaultR().Success(200, "操作成功", "我是数据"))
	})
	group.Get("/response4", func(ctx *msgo.Context) {
		ctx.ErrorHandler(msgo.DefaultR().Fail(500, "操作失败"))
	})
}

// reqResponseTest 请求，返回测试
func reqResponseTest(engine *msgo.Engine) {
	// 主页相关接口
	engine.Group("/").Any("", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%v 您好，首页展示", time.Now())
	})
	engine.Group("/").Any("index", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s index 您好", "李江卫")
	})

	// 用户相关接口
	group := engine.Group("/user")
	group.Use(func(handler msgo.Handler) msgo.Handler {
		return func(ctx *msgo.Context) {
			log.Println("分组执行1前")
			handler(ctx)
			log.Println("分组执行1后")
		}
	}, func(handler msgo.Handler) msgo.Handler {
		return func(ctx *msgo.Context) {
			log.Println("分组执行2前")
			handler(ctx)
			log.Println("分组执行2后")
		}
	})
	group.Head("/1", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s user by any, 您好", "李江卫")
	}, func(handler msgo.Handler) msgo.Handler {
		return func(ctx *msgo.Context) {
			log.Println("方法执行前1")
			handler(ctx)
			log.Println("方法执行后1")
		}
	}, func(handler msgo.Handler) msgo.Handler {
		return func(ctx *msgo.Context) {
			log.Println("方法执行前2")
			handler(ctx)
			log.Println("方法执行后2")
		}
	})
	group.Get("/1", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s user by get, 您好", "李江卫")
	})
	group.Post("/1", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s user by post, 您好", "李江卫")
	})
	group.Put("/1", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s user by put, 您好", "李江卫")
	})
	group.Get("/html", func(ctx *msgo.Context) {
		err := ctx.Html(http.StatusOK, "<h1>李江卫</h1>")
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/html/file", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		err := ctx.HtmlTemplateNoLoad("login.html", template.FuncMap{}, http.StatusOK, dataMap, "tpl/header.html", "tpl/login.html")
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/html/fileGlob", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		err := ctx.HtmlTemplateGlobNoLoad("index.html", template.FuncMap{}, http.StatusOK, dataMap, "tpl/*.html")
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/html/file/loaded", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		err := ctx.HtmlTemplate(http.StatusOK, "index.html", dataMap)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/html/file/loaded/login", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		err := ctx.HtmlTemplate(http.StatusOK, "login.html", dataMap)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/info/json/map", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		err := ctx.JSON(http.StatusOK, dataMap)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/info/xml/map", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		err := ctx.XML(http.StatusOK, dataMap)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/info/json/struct", func(ctx *msgo.Context) {
		user := user{}
		user.Name = "李江卫"
		user.Age = 35
		err := ctx.JSON(http.StatusOK, user)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/info/xml/struct", func(ctx *msgo.Context) {
		user := user{}
		user.Name = "李江卫"
		user.Age = 35
		err := ctx.XML(http.StatusOK, user)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/file", func(ctx *msgo.Context) {
		ctx.File("excel/李江卫.xlsx")
	})

	group.Get("/fileAttachment", func(ctx *msgo.Context) {
		ctx.FileAttachment("excel/李江卫.xlsx", "李江卫.xlsx")
	})

	group.Get("/fileFormFs", func(ctx *msgo.Context) {
		ctx.FileFormFs("李江卫.xlsx", http.Dir("excel"))
	})

	group.Get("/redirect", func(ctx *msgo.Context) {
		ctx.Redirect(http.StatusFound, "/user/info/json/struct")
	})
	group.Get("/string", func(ctx *msgo.Context) {
		ctx.String(http.StatusOK, "%s 您好", "李江卫")
	})
	group.Get("/stringNoFormat", func(ctx *msgo.Context) {
		ctx.String(http.StatusOK, "李江卫您好")
	})

	// 抽象写法
	group.Get("/stringOptions", func(ctx *msgo.Context) {
		ctx.StringOptions(http.StatusOK, "%s 您好", "李江卫")
	})
	group.Get("/stringOptionsNoFormat", func(ctx *msgo.Context) {
		ctx.StringOptions(http.StatusOK, "李江卫您好")
	})
	group.Get("/xmlOptions/struct", func(ctx *msgo.Context) {
		user := user{}
		user.Name = "李江卫"
		user.Age = 35
		err := ctx.XmlOptions(http.StatusOK, user)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/jsonOptions/struct", func(ctx *msgo.Context) {
		user := user{}
		user.Name = "李江卫"
		user.Age = 35
		err := ctx.JsonOptions(http.StatusOK, user)
		if err != nil {
			log.Println("执行异常：", err)
		}
	})
	group.Get("/redirectOptions", func(ctx *msgo.Context) {
		ctx.RedirectOptions(http.StatusFound, "/user/info/json/struct")
	})
	group.Get("/htmlOptions", func(ctx *msgo.Context) {
		ctx.HtmlOptions(http.StatusOK, "<h1>x386</h1>")
	})
	group.Get("/htmlOptionsFile", func(ctx *msgo.Context) {
		dataMap := make(map[string]any)
		dataMap["Name"] = "李江卫"
		ctx.HtmlTemplateOptions(http.StatusOK, "index.html", dataMap)
	})

	group.Put("/*/get", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s user by **/get, 您好", "李江卫")
	})

	paraGroup := engine.Group("/para")
	paraGroup.Get("/info", func(ctx *msgo.Context) {
		name := ctx.GetQuery("name")
		ctx.StringOptions(http.StatusOK, "name is :%s", name)
	})
	paraGroup.Get("/GetArray", func(ctx *msgo.Context) {
		name := ctx.GetArray("name")
		ctx.StringOptions(http.StatusOK, "name is :%v", name)
	})
	paraGroup.Get("/GetQueryArray", func(ctx *msgo.Context) {
		name, ok := ctx.GetQueryArray("name")
		if !ok {
			ctx.String(http.StatusOK, "GetQueryArray by error")
			return
		}
		ctx.StringOptions(http.StatusOK, "name is :%v", name)
	})
	paraGroup.Get("/GetMap", func(ctx *msgo.Context) {
		userMap, ok := ctx.GetMap("user")
		if !ok {
			ctx.String(http.StatusOK, "GetMap by error")
			return
		}
		ctx.JsonOptions(http.StatusOK, userMap)
	})
	paraGroup.Post("/GetForm", func(ctx *msgo.Context) {
		user := ctx.GetForm("user")
		ctx.StringOptions(http.StatusOK, user)
	})
	paraGroup.Post("/GetFormMap", func(ctx *msgo.Context) {
		address, ok := ctx.GetFormMap("address")
		if !ok {
			ctx.StringOptions(http.StatusOK, "GetFormMap error")
			return
		}
		ctx.JsonOptions(http.StatusOK, address)
	})
	paraGroup.Post("/GetFormArrayVal", func(ctx *msgo.Context) {
		friends := ctx.GetFormArrayVal("friends")
		ctx.JsonOptions(http.StatusOK, friends)
	})
	paraGroup.Post("/GetFormArray", func(ctx *msgo.Context) {
		friends, ok := ctx.GetFormArray("friends")
		if !ok {
			ctx.StringOptions(http.StatusOK, "GetFormArray error")
			return
		}
		ctx.JsonOptions(http.StatusOK, friends)
	})
	paraGroup.Post("/file", func(ctx *msgo.Context) {
		file, err := ctx.GetFormFile("file")
		if err != nil {
			ctx.StringOptions(http.StatusInternalServerError, "file get error")
			return
		}
		src, err := file.Open()
		if err != nil {
			ctx.StringOptions(http.StatusInternalServerError, "上传打开异常")
			return
		}
		defer src.Close()

		s := ""
		if i := strings.IndexByte(file.Filename, '.'); i >= 1 {
			s = file.Filename[i:]
		}
		fileName := "upload/" + time.Now().Format("20060102150405") + s
		des, err := os.Create(fileName)
		if err != nil {
			ctx.StringOptions(http.StatusInternalServerError, "创建文件失败")
			return
		}
		defer des.Close()

		buf := make([]byte, 1024)
		_, err = io.CopyBuffer(des, src, buf)
		if err != nil {
			ctx.StringOptions(http.StatusInternalServerError, "保存文件失败")
			return
		}

		ctx.StringOptions(http.StatusOK, fileName)
	})
	paraGroup.Post("/fileUpload", func(ctx *msgo.Context) {
		fileName, err := ctx.UploadFile("file", "upload")
		if err != nil {
			ctx.StringOptions(http.StatusInternalServerError, "上传文件失败")
			return
		}

		ctx.StringOptions(http.StatusOK, fileName)
	})
	paraGroup.Post("/GetJson", func(ctx *msgo.Context) {
		u := &user{}
		err := ctx.BindJson(u)
		if err != nil {
			log.Println(err)
			ctx.StringOptions(http.StatusInternalServerError, "解析json失败")
			return
		}

		ctx.JsonOptions(http.StatusOK, u)
	})
	// 前台多传
	paraGroup.Post("/GetJsonValidate", func(ctx *msgo.Context) {
		ctx.DisallowUnknownFields = true
		u := &user{}
		err := ctx.BindJson(u)
		if err != nil {
			log.Println(err)
			ctx.StringOptions(http.StatusInternalServerError, "解析json失败"+err.Error())
			return
		}

		ctx.JsonOptions(http.StatusOK, u)
	})

	paraGroup.Post("/GetJsonValidateStruct", func(ctx *msgo.Context) {
		ctx.IsValidate = true
		u := &user{}
		err := ctx.BindJson(u)
		if err != nil {
			log.Println(err)
			ctx.StringOptions(http.StatusInternalServerError, "解析json失败"+err.Error())
			return
		}

		ctx.JsonOptions(http.StatusOK, u)
	})
	// 前台多传会报异常；以后台结构体中的属性进行校验
	paraGroup.Post("/GetJsonValidateStructArray", func(ctx *msgo.Context) {
		ctx.DisallowUnknownFields = true
		ctx.IsValidate = true
		u := make([]user, 0)
		err := ctx.BindJson(&u)
		if err != nil {
			log.Println(err)
			ctx.StringOptions(http.StatusInternalServerError, "解析json失败,msg:%s", err.Error())
			return
		}

		ctx.JsonOptions(http.StatusOK, u)
	})

	// 前台多传会报异常；以后台结构体中的属性进行校验
	paraGroup.Post("/GetJsonValidateStructArrayExt", func(ctx *msgo.Context) {
		ctx.DisallowUnknownFields = true
		//ctx.IsValidate = true
		u := make([]user, 0)
		err := ctx.BindJson(&u)
		if err != nil {
			log.Println(err)
			ctx.StringOptions(http.StatusInternalServerError, "解析json失败,msg:%s", err.Error())
			return
		}

		ctx.JsonOptions(http.StatusOK, u)
	})

	// 以后台结构体中的属性进行校验
	paraGroup.Post("/BindXml", func(ctx *msgo.Context) {
		u := make([]user, 0)
		err := ctx.BindXml(&u)
		if err != nil {
			log.Println(err)
			ctx.StringOptions(http.StatusInternalServerError, "解析xml失败,msg:%s", err.Error())
			return
		}

		ctx.XmlOptions(http.StatusOK, u)
	})
}

// loggintTest 非日志组件测试
func loggintTest(engine *msgo.Engine) {
	// 日志测试
	logGroup := engine.Group("/log")
	logGroup.Use(msgo.Logging)
	logGroup.Get("/info", func(ctx *msgo.Context) {
		//log.Println("log info handler")
		ctx.StringOptions(http.StatusOK, "ok")
	})

	// 日志全框架
	extFields := make(map[string]any, 0)
	logger := logs.Default().WithField(extFields).WithFormat(&logs.JsonFormatter{}) //.WithFormat(&logs.TextFormatter{})
	logger.SetPath("./logs")
	defer func() {
		logger.Close()
	}()
	logPlugGroup := engine.Group("/logPlug")
	logPlugGroup.Get("/test", func(ctx *msgo.Context) {
		ctx.StringOptions(http.StatusOK, "ok")
		extFields["user"] = "李江卫"
		extFields["age"] = 35
		logger.Debug("testing  debug 日志")
		logger.Info("testing info 日志")
		logger.Error("testing error 日志")
	})
}
