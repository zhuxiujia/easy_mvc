package main

import (
	"github.com/zhuxiujia/easy_mvc"
	"github.com/zhuxiujia/easy_mvc/easy_swagger"
	_ "github.com/zhuxiujia/easy_mvc/easy_swagger/dist/statik"
	"net/http"
)

//type Swagger struct {
//	paths string `json:"paths"`
//}
type TestController struct {
	easy_mvc.Controller `path:"/api"`
	//登录接口案例,返回值默认转json，如果要返回其他东西，请在参数里加上 request *http.Request 把content-type 改了，然后可以自行处理（或者直接兼容标准库func(writer http.ResponseWriter, request *http.Request)）
	Login func(phone string, pwd string, age *int) interface{} `path:"/login" arg:"phone,pwd,age" doc_arg:"phone:手机号,pwd:密码,age:年龄" doc:"登录接口"`
	//兼容go标准库http案例,可以无返回值
	Login2 func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r"`
	Login3 func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r" method:"get"`
	Login4 func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r"`

	UserInfo func() interface{} `path:"/api/login2"`
}

func main() {
	var controller = TestController{}
	controller.Init(&controller) //初始化

	easy_swagger.EnableSwagger("localhost:9993", easy_swagger.SwaggerConfig{})
	http.ListenAndServe("localhost:9993", nil)
}
