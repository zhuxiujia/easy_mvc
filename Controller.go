package easy_mvc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

const ContentType = "Content-Type"

type HttpChan struct {
	Func func(path string, w http.ResponseWriter, r *http.Request) bool //函数返回error则中断执行
	Name string
}

type HttpErrorHandle struct {
	Func func(err interface{}, w http.ResponseWriter, r *http.Request)
	Name string
}

type HttpResultHandle struct {
	Func func(result *interface{}, w http.ResponseWriter, r *http.Request) bool
	Name string
}

//全局http调用链，过滤器,return error 不为nil则不继续执行
var GlobalHttpChan = []*HttpChan{}

//全局错误调用链
var GlobalErrorHandleChan = []*HttpErrorHandle{}

var GlobalResultHandleChan = []*HttpResultHandle{}

//全局错误处理器
var GlobalErrorHandle = func(w http.ResponseWriter, r *http.Request) {
	// 发生宕机时，获取panic传递的上下文并打印
	err := recover()
	for _, itemFunc := range GlobalErrorHandleChan {
		if itemFunc != nil {
			itemFunc.Func(err, w, r)
		}
	}
}

//注册全局错误
func RegisterGlobalResultHandleChan(handle *HttpResultHandle) {
	if handle == nil {
		return
	}
	if handle.Name == "" {
		panic("[easy_mvc] handle must have a name!")
	}
	for index, v := range GlobalResultHandleChan {
		if v.Name == handle.Name {
			GlobalResultHandleChan[index] = v
		}
	}
	GlobalResultHandleChan = append(GlobalResultHandleChan, handle)
}

//注册全局错误
func RegisterGlobalErrorHandleChan(handle *HttpErrorHandle) {
	if handle.Name == "" {
		panic("[easy_mvc] handle must have a name!")
	}
	for index, v := range GlobalErrorHandleChan {
		if v.Name == handle.Name {
			GlobalErrorHandleChan[index] = v
		}
	}
	GlobalErrorHandleChan = append(GlobalErrorHandleChan, handle)
}

//注册http调用链/过滤器
func RegisterGlobalHttpChan(handle *HttpChan) {
	if handle.Name == "" {
		panic("[easy_mvc] handle must have a name!")
	}
	for index, v := range GlobalHttpChan {
		if v.Name == handle.Name {
			GlobalHttpChan[index] = v
		}
	}
	GlobalHttpChan = append(GlobalHttpChan, handle)
}

func init() {
	var defHttpHandle = HttpChan{
		Func: func(path string, w http.ResponseWriter, r *http.Request) bool {
			w.Header().Set(ContentType, "application/json") //框架默认使用json处理结果
			return false
		},
		Name: "DefHttpHandle",
	}
	GlobalHttpChan = append(GlobalHttpChan, &defHttpHandle)

	var defHttpErrorHandle = HttpErrorHandle{
		Func: func(err interface{}, w http.ResponseWriter, r *http.Request) {
			if err != nil {
				switch err.(type) {
				case runtime.Error: // 运行时错误
					log.Println("runtime error:", err)
				default: // 非运行时错误
					log.Println("error:", err)
				}
			}
		},
		Name: "DefHttpErrorHandle",
	}

	GlobalErrorHandleChan = append(GlobalErrorHandleChan, &defHttpErrorHandle)
}

//例如 SendSms(writer http.ResponseWriter, request *http.Request)  `path:"/" arg:"w,r" `
//模板 `path:"" arg:"" `
type Controller struct {
}

var targetPaths = map[string]func(w http.ResponseWriter, r *http.Request){}
var pathMethods = map[string]bool{}

func (it *Controller) Init(arg interface{}) {
	var argV = reflect.ValueOf(arg)
	if argV.Kind() != reflect.Ptr {
		panic("[easy_mvc] Init value " + argV.String() + " must be struct{} ptr!")
	}
	for {
		if argV.Kind() == reflect.Ptr {
			argV = argV.Elem()
		} else {
			break
		}
	}
	if argV.Type().Kind() != reflect.Struct {
		panic("[easy_mvc] Init value must be a struct{} ptr!")
	}

	var argType = argV.Type()
	var rootPath = checkHaveRootPath(argType)
	for i := 0; i < argType.NumField(); i++ {
		var funcField = argType.Field(i)
		var field = argV.Field(i)
		if funcField.Type.Kind() != reflect.Func {
			continue
		}
		if funcField.Type.NumOut() > 1 {
			log.Println("[easy_rpc] warning not registed !============= " + funcField.Name + " " + funcField.Type.String() + " func return num > 1 ")
			continue
		}
		var tagPath = rootPath + funcField.Tag.Get("path")
		if tagPath == "" {
			continue
		}
		var tagArg = funcField.Tag.Get("arg")
		var tagArgs []string
		if tagArg != "" {
			tagArgs = strings.Split(tagArg, ",")
		} else {
			tagArgs = []string{}
		}
		//check len
		if len(tagArgs) != funcField.Type.NumIn() {
			panic("[easy_mvc] " + argType.String() + "." + funcField.Name + "() args.len(" + fmt.Sprint(funcField.Type.NumIn()) + ") != tag arg.len(" + fmt.Sprint(len(tagArgs)) + ")!")
		}
		//反射方法类型
		var funSplits = [][]string{}
		var funInTypes = []reflect.Type{}
		for i := 0; i < funcField.Type.NumIn(); i++ {
			funInTypes = append(funInTypes, funcField.Type.In(i))
			var defs = strings.Split(tagArgs[i], ":")
			funSplits = append(funSplits, defs)
		}
		var method = funcField.Tag.Get("method")
		//check repeat method
		var have = pathMethods[tagPath+method]
		if have == true {
			panic(fmt.Sprint("[easy_mvc] find repeat method of ", tagPath))
		}
		pathMethods[tagPath+method] = true
		//decode http func
		var httpFunc = func(w http.ResponseWriter, r *http.Request) {
			//default param
			defer GlobalErrorHandle(w, r)
			//chan
			for _, v := range GlobalHttpChan {
				if v.Func(tagPath, w, r) {
					return
				}
			}

			if method != "" {
				if !strings.EqualFold(r.Method, method) {
					w.Write([]byte("[easy_mvc] http method not allow! current use:" + r.Method + "you must use:" + method))
					return
				}
			}

			var args = []reflect.Value{}
			for i := 0; i < len(funInTypes); i++ {
				var argItemType = funInTypes[i]
				var defs = funSplits[i]

				var httpArg = "" //http arg
				//JsonArg
				var req_content = r.Header.Get(ContentType)
				if req_content == "application/json" {
					b, err := ioutil.ReadAll(r.Body)
					if err != nil {
						var errStr = "  error = " + err.Error()
						w.Write([]byte("[easy_mvc] parser http arg fail: path=" + tagPath + ",type=" + argItemType.String() + ",tag=" + tagArgs[i] + ",error=" + errStr))
						return
					}
					httpArg = string(b)
				} else {
					r.ParseForm()
					httpArg = r.Form.Get(defs[0]) //http arg
				}
				//FormArg
				var convertV, e = convert(httpArg, argItemType, defs[0], w, r)
				if convertV.IsValid() && e == nil {
					if len(defs) == 2 {
						if argItemType.Kind() == reflect.Ptr && convertV.IsNil() {
							convertV, e = convert(defs[1], argItemType, defs[0], w, r)
							if e != nil {
								var errStr = "  error = " + e.Error()
								w.Write([]byte("[easy_mvc] parser http arg fail: path=" + tagPath + ",type=" + argItemType.String() + ",tag=" + tagArgs[i] + ",error=" + errStr))
								return
							}
						} else if argItemType.Kind() == reflect.String {
							convertV = reflect.ValueOf(defs[1])
						}
					}
					args = append(args, convertV)
				} else {
					var convertSuccess *reflect.Value
					if len(defs) == 2 {
						convertV, e = convert(defs[1], argItemType, defs[0], w, r)
						if e != nil {
							var errStr = ""
							if e != nil {
								errStr = "  error = " + e.Error()
							}
							w.Write([]byte("[easy_mvc] parser http arg fail: path=" + tagPath + ",type=" + argItemType.String() + ",tag=" + tagArgs[i] + ",error=" + errStr))
							return
						} else {
							convertSuccess = &convertV
						}
					}
					if argItemType.Kind() == reflect.Ptr || convertSuccess != nil {
						args = append(args, convertV)
						continue
					}
					var errStr = ""
					if e != nil {
						errStr = "  error = " + e.Error()
					}
					w.Write([]byte("[easy_mvc] parser http arg fail: path=" + tagPath + ",type=" + argItemType.String() + ",tag=" + tagArgs[i] + ",error=" + errStr))
					return
				}
			}
			var results = field.Call(args)
			var contentType = w.Header().Get("Content-Type")
			if results != nil && len(results) > 0 {
				switch contentType {
				case "application/json":
					var result = results[0].Interface()
					for _, item := range GlobalResultHandleChan {
						if item.Func(&result, w, r) { //success return,else next
							return
						}
					}
					var b, _ = json.Marshal(result)
					w.Write(b)
					break
				}
			}
		}
		log.Println("[easy_mvc] http.HandleFunc " + argType.String() + "  =>  " + funcField.Name + " " + funcField.Type.String() + strings.Replace(string(" "+funcField.Tag), funcField.Tag.Get("path"), tagPath, -1))
		var oldPath = targetPaths[tagPath]
		if oldPath == nil {
			oldPath = func(w http.ResponseWriter, r *http.Request) {
				httpFunc(w, r)
			}
		} else {
			oldPath = func(w http.ResponseWriter, r *http.Request) {
				oldPath(w, r)
				httpFunc(w, r)
			}
		}
		targetPaths[tagPath] = oldPath
	}
	for path, _ := range targetPaths {
		http.HandleFunc(path, targetPaths[path])
	}
	//存入上下文
	ControllerTable.Store(argType.Name(), arg)
}

func checkHaveRootPath(argType reflect.Type) string {
	for i := 0; i < argType.NumField(); i++ {
		var field = argType.Field(i)
		if field.Type.String() == "easy_mvc.Controller" {
			var rootPath = field.Tag.Get("path")
			return rootPath
		}
	}
	return ""
}
