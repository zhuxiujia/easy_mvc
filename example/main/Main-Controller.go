package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zhuxiujia/easy_mvc"
	"github.com/zhuxiujia/easy_mvc/easy_swagger"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type TestUserVO struct {
	Name string
}

type TestController struct {
	easy_mvc.Controller `path:"/api"`
	//登录接口案例,返回值默认转json，如果要返回其他东西，请在参数里加上 request *http.Request 把content-type 改了，然后可以自行处理（或者直接兼容标准库func(writer http.ResponseWriter, request *http.Request)）
	Login func(phone string, pwd string, age *int) interface{} `method:"get" path:"/login" arg:"phone,pwd,age" doc:"登录接口" doc_arg:"phone:手机号,pwd:密码,age:年龄"`
	//兼容go标准库http案例,可以无返回值
	Login2 func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r" doc:"登录接口"`
	Login3 func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r" method:"get" doc:"登录接口"`
	Login4 func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone:18969542172,pwd,r" doc:"登录接口"`
	Upload func(file easy_mvc.MultipartFile) interface{}                       `path:"/upload" arg:"file" doc:"文件上传"`
	Json   func(js string) interface{}                                         `path:"/json" arg:"js" doc:"json数据,需要Header/Content-Type设置application/json"`

	UserInfo  func() interface{}                                   `path:"/api/login2" method:"post"`
	UserInfo2 func(name string, request *http.Request) interface{} `path:"/api/login2/{name}" method:"get" arg:"name,r"` //path参数
}

func (it *TestController) New(router *mux.Router) {
	it.Login = func(phone string, pwd string, age *int) interface{} {
		var ageStr = ""
		if age != nil {
			ageStr = fmt.Sprint(*age)
		} else {
			ageStr = "nil"
		}
		return fmt.Sprint("do Login phone string, pwd string, age *int :", phone, ",", pwd, ",", ageStr)
	}
	it.UserInfo = func() interface{} {
		return TestUserVO{Name: "UserInfo"}
	}
	it.UserInfo2 = func(name string, r *http.Request) interface{} {
		//vars := mux.Vars(r)
		//name := vars["name"]
		return TestUserVO{Name: "UserInfo2,name=" + name}
	}
	it.Login2 = func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Login2"))
	}
	it.Login3 = func(writer http.ResponseWriter, request *http.Request) interface{} {
		writer.Write([]byte("Login3"))
		return nil
	}
	it.Login4 = func(phone string, pwd string, request *http.Request) interface{} {
		return "Login4:" + phone
	}
	it.Upload = func(file easy_mvc.MultipartFile) interface{} {
		if file.Error != nil {
			log.Println("upload success=============", file.Error)
			return "fail"
		}
		log.Println("upload success=============" + file.Filename)
		return "success"
	}

	it.Json = func(js string) interface{} {
		var m = map[string]interface{}{}
		json.Unmarshal([]byte(js), &m)
		for k, v := range m {
			println("json_key:", k)
			println("json_value:", fmt.Sprint(v))
		}
		return js
	}

	it.Init(&it, router) //必须初始化，而且是指针
}

func main() {
	//first define router
	//首先，初始化路由
	var router = mux.Router{}
	http.Handle("/", &router)

	//自定义一个全局错误处理器 到调用链中
	easy_mvc.RegisterGlobalErrorHandleChan(&easy_mvc.HttpErrorHandle{
		Func: func(err interface{}, w http.ResponseWriter, r *http.Request) {
			if err != nil {
				println(err.(error).Error())
				w.Write([]byte(err.(error).Error()))
			}
		},
		Name: "ErrorFilter",
	})

	easy_mvc.RegisterGlobalResultHandleChan(&easy_mvc.HttpResultHandle{
		Func: func(result *interface{}, w http.ResponseWriter, r *http.Request) bool {
			var v = reflect.ValueOf(*result)
			if strings.Contains(v.Type().String(), "error") {
				*result = TestUserVO{}
			}
			return false
		},
		Name: "ResultFilter",
	})

	//初始化 控制器
	var testController = TestController{}
	testController.New(&router)

	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("this is /"))
	})

	//http.HandleFunc("/doc", func(writer http.ResponseWriter, request *http.Request) {
	//	writer.Write(easy_swagger.ScanControllerContext(easy_swagger.SwaggerConfig{}))
	//})

	//启用swagger ui 前端界面
	easy_swagger.EnableSwagger("localhost:8080", easy_swagger.SwaggerConfig{
		//SecurityDefinitionConfig: &easy_swagger.SecurityDefinitionConfig{
		//	easy_swagger.SecurityDefinition{
		//		ApiKey: easy_swagger.ApiKey{
		//			Type: "apiKey",
		//			Name: "access_token",
		//			In:   "query",
		//		},
		//	},
		//	"/api/login2",
		//},
	})

	println("服务启动于 ", "127.0.0.1:8080")
	//使用标准库启动http
	http.ListenAndServe("127.0.0.1:8080", nil)
}
