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

//全局http调用链，过滤器,return error 不为nil则不继续执行
var GlobalHttpChan = []func(w http.ResponseWriter, r *http.Request) error{}

//全局错误调用链
var GlobalErrorHandleChan = []func(err interface{}){}

//全局错误处理器
var GlobalErrorHandle = func() {
	// 发生宕机时，获取panic传递的上下文并打印
	err := recover()
	switch err.(type) {
	case runtime.Error: // 运行时错误
		log.Println("runtime error:", err)
	default: // 非运行时错误
		log.Println("error:", err)
	}
	for _, itemFunc := range GlobalErrorHandleChan {
		if itemFunc != nil {
			itemFunc(err)
		}
	}
}

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
		//decode http func
		var httpFunc = func(w http.ResponseWriter, r *http.Request) {

			defer GlobalErrorHandle()
			//chan
			for _, v := range GlobalHttpChan {
				var e = v(w, r)
				if e != nil {
					return
				}
			}
			//default param
			r.ParseForm()
			var args = []reflect.Value{}
			for i := 0; i < funcField.Type.NumIn(); i++ {
				var argItemType = funcField.Type.In(i)
				var defs = strings.Split(tagArgs[i], ":")
				var httpArg = r.Form.Get(defs[0]) //http arg
				var convertV, e = convert(httpArg, argItemType, w, r)
				if convertV.IsValid() && e == nil || len(defs) == 2 {
					if convertV.IsValid() && e == nil {
						args = append(args, convertV)
					} else {
						var convertV, e = convert(defs[1], argItemType, w, r)
						if e != nil {
							var errStr = ""
							if e != nil {
								errStr = "  error = " + e.Error()
							}
							w.Write([]byte("[easy_mvc] parser http arg fail:" + argItemType.String() + ":" + tagArgs[i] + errStr))
							return
						}
						args = append(args, convertV)
					}
				} else {
					var errStr = ""
					if e != nil {
						errStr = "  error = " + e.Error()
					}
					w.Write([]byte("[easy_mvc] parser http arg fail:" + argItemType.String() + ":" + tagArgs[i] + errStr))
					return
				}
			}
			var results = field.Call(args)
			if results != nil && len(results) > 0 {
				var b, _ = json.Marshal(results[0].Interface())
				w.Write(b)
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
