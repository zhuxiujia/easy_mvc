package easy_mvc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type HttpChan struct {
	Func func(w http.ResponseWriter, r *http.Request) bool //函数返回error则中断执行
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
		Func: func(w http.ResponseWriter, r *http.Request) bool {
			w.Header().Set("Content-type", "application/json") //框架默认使用json处理结果
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
	check_null_str string
}

func (it *Controller) Init(arg interface{}) {
	var argType = reflect.TypeOf(arg)
	if argType.Kind() != reflect.Ptr {
		panic("[easy_mvc] Init value " + argType.String() + " must be a ptr!")
	}
	argType = argType.Elem()
	var v = reflect.ValueOf(arg).Elem()
	var rootPath = checkHaveRootPath(argType)
	for i := 0; i < argType.NumField(); i++ {
		var funcField = argType.Field(i)
		var field = v.Field(i)
		if funcField.Type.Kind() != reflect.Func {
			continue
		}
		if funcField.Type.NumOut() > 1 {
			log.Println("[easy_rpc] warning ============= " + funcField.Name + " " + funcField.Type.String() + " not registed to http!")
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

		//decode http func
		var httpFunc = func(w http.ResponseWriter, r *http.Request) {
			//default param
			r.ParseForm()
			defer GlobalErrorHandle(w, r)
			//chan
			for _, v := range GlobalHttpChan {
				if v.Func(w, r) {
					return
				}
			}
			var args = []reflect.Value{}
			for i := 0; i < len(funInTypes); i++ {
				var argItemType = funInTypes[i]
				var defs = funSplits[i]
				var httpArg = r.Form.Get(defs[0]) //http arg
				var convertV, e = convert(httpArg, argItemType, w, r)
				if convertV.IsValid() && e == nil {
					if len(defs) == 2 && convertV.IsNil() {
						convertV, e = convert(defs[1], argItemType, w, r)
						if e != nil {
							var errStr = ""
							if e != nil {
								errStr = "  error = " + e.Error()
							}
							w.Write([]byte("[easy_mvc] parser http arg fail:" + argItemType.String() + ":" + tagArgs[i] + errStr))
							return
						}
					}
					args = append(args, convertV)
				} else {
					if len(defs) == 2 {
						convertV, e = convert(defs[1], argItemType, w, r)
						if e != nil {
							var errStr = ""
							if e != nil {
								errStr = "  error = " + e.Error()
							}
							w.Write([]byte("[easy_mvc] parser http arg fail:" + argItemType.String() + ":" + tagArgs[i] + errStr))
							return
						}
					}
					if argItemType.Kind() == reflect.Ptr {
						args = append(args, convertV)
						continue
					}
					var errStr = ""
					if e != nil {
						errStr = "  error = " + e.Error()
					}
					w.Write([]byte("[easy_mvc] parser http arg fail:" + argItemType.String() + ":" + tagArgs[i] + errStr))
					return
				}
			}
			var results = field.Call(args)
			var contentType = w.Header().Get("Content-type")
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
		log.Println("[easy_mvc] http Handle " + funcField.Name + " " + funcField.Type.String() + string(" "+funcField.Tag))
		http.HandleFunc(tagPath, httpFunc)
	}

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

func convert(value string, tItemTypeFieldType reflect.Type, w http.ResponseWriter, r *http.Request) (reflect.Value, error) {
	if tItemTypeFieldType.Kind() == reflect.String {
		return reflect.ValueOf(value), nil
	} else {
		if tItemTypeFieldType.Kind() == reflect.Bool {
			newValue, e := strconv.ParseBool(value)
			if e != nil {
				return reflect.Value{}, e
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Int || tItemTypeFieldType.Kind() == reflect.Int16 || tItemTypeFieldType.Kind() == reflect.Int32 || tItemTypeFieldType.Kind() == reflect.Int64 {
			newValue, e := strconv.ParseInt(value, 0, 64)
			if e != nil {
				return reflect.Value{}, e
			}
			if tItemTypeFieldType.Kind() == reflect.Int {
				return reflect.ValueOf(int(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Int16 {
				return reflect.ValueOf(int16(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Int32 {
				return reflect.ValueOf(int32(newValue)), nil
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Uint || tItemTypeFieldType.Kind() == reflect.Uint8 || tItemTypeFieldType.Kind() == reflect.Uint16 || tItemTypeFieldType.Kind() == reflect.Uint32 || tItemTypeFieldType.Kind() == reflect.Uint64 {
			newValue, e := strconv.ParseUint(value, 0, 64)
			if e != nil {
				return reflect.Value{}, e
			}
			if tItemTypeFieldType.Kind() == reflect.Uint {
				return reflect.ValueOf(uint(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Uint16 {
				return reflect.ValueOf(uint16(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Uint32 {
				return reflect.ValueOf(uint32(newValue)), nil
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Float32 || tItemTypeFieldType.Kind() == reflect.Float64 {
			newValue, e := strconv.ParseFloat(value, 64)
			if e != nil {
				return reflect.Value{}, e
			}
			if tItemTypeFieldType.Kind() == reflect.Float32 {
				return reflect.ValueOf(float32(newValue)), nil
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Struct {
			if tItemTypeFieldType.String() == "time.Time" {
				newValue, e := time.Parse(string(time.RFC3339), value)
				if e != nil {
					return reflect.Value{}, e
				}
				return reflect.ValueOf(newValue), nil
			} else {

			}
		} else if tItemTypeFieldType.Kind() == reflect.Interface {
			if tItemTypeFieldType.String() == "http.ResponseWriter" {
				return reflect.ValueOf(w), nil
			}
		} else if tItemTypeFieldType.Kind() == reflect.Ptr {
			if tItemTypeFieldType.String() == "*http.Request" {
				return reflect.ValueOf(r), nil
			}
			if value == "" {
				return reflect.Zero(tItemTypeFieldType), nil
			}
			var v, e = convert(value, tItemTypeFieldType.Elem(), w, r)
			var newPtrV = reflect.New(tItemTypeFieldType.Elem())
			if v.IsValid() {
				newPtrV.Elem().Set(v)
			}
			return newPtrV, e
		}
	}

	return reflect.Value{}, nil
}
