#### GO mvc框架，支持IOC,AOP,DI 基于golang标准库,http库和Tag机制实现.免生成中间代码，非常容易使用
#### 这可能是你能找到的最灵活，最方便的框架
## 核心特性

* 整体基于反射tag，所有配置（包括http方法，路径，参数，swagger文档参数）都集中于你定义的函数之后
* 轻量 完全兼容标准库的http，意味着和标准库一般稳定，可以混合搭配使用，扩展性极高
* 拦截器 支持（例如非常方便的检查用户登录，提取用户登录数据，支持JWT token，Oath2Token更加方便的接入）
* 过滤器 支持
* 全局错误处理器链 支持
* 使用tag 定义 http请求参数，包含 *int,*string,*float 同时支持标准库的 writer http.ResponseWriter, request *http.Request
* Json参数支持（app端上传时需要Header，Content-Type设置为application/json）
* 支持参数默认值 只需在tag中 定义，例如 func(phone string, pwd string, age *int) interface{} arg:"phone,pwd,age:1"  其中 arg没有传参则默认为1
* 指针参数可为空（nil）非指针参数 如果没有值框架会拦截
* root path支持，类似spring controller定义一个基础的path加控制器具体方法的http path
* 支持swagger ui 动态文档，免生成任何中间go文件 基于Tag和反射实现的swagger动态文档


## Controller
``` go
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
	easy_mvc.Controller `path:"/api"`         //基路由
	//登录接口案例,返回值默认转json，如果要返回其他东西，请在参数里加上 request *http.Request 把content-type 改了，然后可以自行处理（或者直接兼容标准库func(writer http.ResponseWriter, request *http.Request)）
	Login func(phone string, pwd string, age *int) interface{} `path:"/login" arg:"phone,pwd,age" doc_arg:"phone:手机号,pwd:密码,age:年龄"`
	//兼容go标准库http案例,可以无返回值
	Login2 func(writer http.ResponseWriter, request *http.Request)             `path:"/login2" arg:"w,r"`
	Login3 func(writer http.ResponseWriter, request *http.Request) interface{} `path:"/login3" arg:"w,r" method:"get"`
	Login4 func(phone string, pwd string, request *http.Request) interface{}   `path:"/login4" arg:"phone,pwd,r"`
	Json   func(js string) interface{}                                         `path:"/json" arg:"js" doc:"json数据,需要Header/Content-Type设置application/json"`

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

	println("服务启动于 ","127.0.0.1:8080")

	//使用标准库启动http
	http.ListenAndServe("127.0.0.1:8080", nil)
}
```

* 启动Main-Controller.go 后即可在日志查看 具体挂载的http接口地址，例如上面定义的/login2 挂载为(基地址+具体接口地址) 如 /api/login2
* 也可以执行Main-Swagger.go,即可在localhost:9993 查看swagger ui接口地址
``` log
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  Login func(string, string, *int) interface {} method:"get" path:"/api/login" arg:"phone,pwd,age" doc:"登录接口" doc_arg:"phone:手机号,pwd:密码,age:年龄"
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  Login2 func(http.ResponseWriter, *http.Request) path:"/api/login2" arg:"w,r" doc:"登录接口"
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  Login3 func(http.ResponseWriter, *http.Request) interface {} path:"/api/login3" arg:"w,r" method:"get" doc:"登录接口"
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  Login4 func(string, string, *http.Request) interface {} path:"/api/login4" arg:"phone:18969542172,pwd,r" doc:"登录接口"
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  Upload func(easy_mvc.MultipartFile) interface {} path:"/api/upload" arg:"file" doc:"文件上传"
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  Json func(string) interface {} path:"/api/json" arg:"js" doc:"json数据,需要Header/Content-Type设置application/api/json"
2020/02/22 00:42:55 [easy_mvc] http.HandleFunc main.TestController  =>  UserInfo func() interface {} path:"/api/api/login2"
2020/02/22 00:42:55 [easy_rpc] warning not registed !============= UserInfo2 func() (interface {}, interface {}) func return num > 1 
服务启动于  127.0.0.1:8080
```
