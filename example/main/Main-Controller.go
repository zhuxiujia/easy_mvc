package main

import (
	"encoding/json"
	"fmt"
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

	UserInfo  func() interface{}                `path:"/api/login2"`
	UserInfo2 func() (interface{}, interface{}) `path:"/api/login2"`
}

func (it *TestController) New() {
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
		return TestUserVO{}
	}
	it.Login2 = func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("fuck"))
	}
	it.Login3 = func(writer http.ResponseWriter, request *http.Request) interface{} {

		return nil
	}
	it.Login4 = func(phone string, pwd string, request *http.Request) interface{} {

		return phone
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
		for k,v := range m {
			println("json_key:",k)
			println("json_value:",fmt.Sprint(v))
		}
		return js
	}

	it.Init(&it) //必须初始化，而且是指针
}

func main() {
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
	testController.New()

	//你也可以使用标准库的api（使用标准库不经过easy_mvc）
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("yes"))
	})

	//可以启动一个swagger api的接口，提供给swagger
	http.HandleFunc("/doc", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(easy_swagger.ScanControllerContext(easy_swagger.SwaggerConfig{}))
	})

	println("服务启动于 ","127.0.0.1:8080")
	//使用标准库启动http
	http.ListenAndServe("127.0.0.1:8080", nil)
}
