package easy_swagger

import (
	"fmt"
	"github.com/zhuxiujia/easy_mvc"
	"net/http"
	"testing"
)

type TestController struct {
	easy_mvc.Controller `path:"/api"`
	//登录接口案例,返回值默认转json，如果要返回其他东西，请在参数里加上 request *http.Request 把content-type 改了，然后可以自行处理（或者直接兼容标准库func(writer http.ResponseWriter, request *http.Request)）
	Login func(phone string, pwd string, age *int) interface{} `path:"/login" arg:"phone,pwd,age" note:"phone:手机号,pwd:密码,age:年龄"`
	//兼容go标准库http案例,可以无返回值
	Login2 func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r"`
	Login3 func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r" method:"get"`
	Login4 func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r"`

	UserInfo  func() interface{}                `path:"/api/login2"`
	UserInfo2 func() (interface{}, interface{}) `path:"/api/login2"`
}

func TestYaml(t *testing.T) {
	var c = TestController{}

	fmt.Printf("--- m dump:\n%s\n\n", string(CreateSwaggerYaml(Scan(&c))))
}
