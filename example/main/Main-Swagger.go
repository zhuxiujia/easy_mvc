package main

import (
	"github.com/gorilla/mux"
	"github.com/zhuxiujia/easy_mvc"
	"github.com/zhuxiujia/easy_mvc/easy_swagger"
	_ "github.com/zhuxiujia/easy_mvc/easy_swagger/dist/statik"
	"net/http"
)

//type Swagger struct {
//	paths string `json:"paths"`
//}
type TestSwaggerController struct {
	easy_mvc.Controller `path:"/api" doc:"asdfsad"`
	//登录接口案例,返回值默认转json，如果要返回其他东西，请在参数里加上 request *http.Request 把content-type 改了，然后可以自行处理（或者直接兼容标准库func(writer http.ResponseWriter, request *http.Request)）
	Login func(phone string, pwd string, age *int) interface{} `path:"/login" arg:"phone,pwd,age" doc_arg:"phone:手机号,pwd:密码,age:年龄" doc:"登录接口"`
	//兼容go标准库http案例,可以无返回值
	Login2 func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r" doc_arg:"w:_,r:_"`
	Login3 func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r" method:"get" doc_arg:"w:_,r:_"`
	Login4 func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r"`

	UserInfo  func(phone string, test bool) interface{}            `path:"/api/login2" method:"post"  arg:"phone,test" doc_arg:"phone:手机号,test:测试bool" doc:"信息接口"`
	UserInfo2 func(name string, request *http.Request) interface{} `path:"/api/login2/{name}" method:"get" arg:"name,r"` //path参数
}

func main() {
	//first define router
	var router = mux.Router{}

	var controller = TestSwaggerController{}
	controller.Init(&controller, &router) //初始化

	easy_swagger.EnableSwagger("localhost:9993", easy_swagger.SwaggerConfig{
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
	http.ListenAndServe("localhost:9993", nil)
}
