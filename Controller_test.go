package easy_mvc

import (
	"net/http"
	"testing"
)

type TestUserVO struct {
	Name string
}

type TestController struct {
	Controller `path:"/api"`
	Login      func(phone string, pwd string, age *int) interface{}                `path:"/login" arg:"phone,pwd,age" note:"phone:手机号,pwd:密码,age:年龄"`
	Login2     func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r"`
	Login3     func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r"`
	Login4     func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r"`

	UserInfo func() interface{} `path:"/api/login2"`
}

func (it TestController) New() TestController {
	it.Login = func(phone string, pwd string, age *int) interface{} {
		println("do Login")
		return phone + pwd
	}
	it.UserInfo = func() interface{} {
		return TestUserVO{}
	}
	it.Login2 = func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("fuck"))
	}
	it.Init(&it)
	return it
}

func TestController_Init(t *testing.T) {
	RegisterGlobalErrorHandleChan(&HttpErrorHandle{
		Func: func(err interface{}, w http.ResponseWriter, r *http.Request) {
			if err != nil {
				println(err.(error).Error())
				w.Write([]byte(err.(error).Error()))
			}
		},
		Name: "custom",
	})

	TestController{}.New()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("yes"))
	})

	http.ListenAndServe("127.0.0.1:8080", nil)
}
