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
	Login      func(phone string, pwd string) interface{}                          `path:"/login" arg:"phone,pwd" method:"get"`
	Login2     func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login2" arg:"phone,pwd" method:"post"`
	Login3     func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"phone,pwd" method:"put"`
	Login4     func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r" method:"delete"`

	UserInfo func() interface{} `path:"/api/login2" rsp:"json"`
}

func (it TestController) New() TestController {

	it.Login = func(phone string, pwd string) interface{} {

		return nil
	}
	it.UserInfo = func() interface{} {

		return TestUserVO{}
	}
	it.Init(&it)
	return it
}

func TestController_Init(t *testing.T) {
	TestController{}.New()
	//Provide("c", &test)

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("yes"))
	})
	http.ListenAndServe("127.0.0.1:8080", nil)
}
