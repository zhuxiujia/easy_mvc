#### GO mvc框架，支持IOC,AOP,DI 基于golang标准库,http库和Tag机制实现.免生成中间代码，非常容易使用

## 核心特性

* 轻量 完全兼容标准库的http，可以混合使用，扩展性机高
* 拦截器 支持（例如非常方便的检查用户登录，提取用户登录数据）
* 过滤器 支持
* 全局错误处理器链 支持
* 使用tag 定义 http请求参数，包含 *int,*string,*float 同时支持标准库的 writer http.ResponseWriter, request *http.Request
* 支持参数默认值 只需在tag中 定义，例如 func(phone string, pwd string, age *int) interface{} arg:"phone,pwd,age:1"  其中 arg没有传参则默认为1
* 指针参数可为空（nil）非指针参数 如果没有值框架会拦截
* root path支持，类似spring controller定义一个基础的path加控制器具体方法的http path
* 基于Tag和反射动态文档,免除繁琐的文档编写和代码生成（即将到来）


```
import (
	"errors"
	"github.com/zhuxiujia/easy_mvc"
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
	Login func(phone string, pwd string, age *int) interface{} `path:"/login" arg:"phone,pwd,age" note:"phone:手机号,pwd:密码,age:年龄"`
	//兼容go标准库http案例,可以无返回值
	Login2 func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r"`
	Login3 func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r" method:"get"`
	Login4 func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r"`

	UserInfo  func() interface{}                `path:"/api/login2"`
}

func (it *TestController) New() {
	it.Login = func(phone string, pwd string, age *int) interface{} {
		println("do Login")
		return errors.New("dsf")
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

	println("启动成功··")
	//使用标准库启动http
	http.ListenAndServe("127.0.0.1:8080", nil)
}
```
