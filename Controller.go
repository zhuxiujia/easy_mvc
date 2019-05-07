package easy_mvc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type Controller struct {
}

func (it *Controller) Init(arg interface{}) {
	var argType = reflect.TypeOf(arg)
	if argType.Kind() != reflect.Ptr {
		panic("[easy_mvc] Init value " + argType.String() + " must be a ptr!")
	}
	argType = argType.Elem()
	var v = reflect.ValueOf(arg).Elem()
	for i := 0; i < argType.NumField(); i++ {
		var funcField = argType.Field(i)
		var field = v.Field(i)
		if funcField.Type.Kind() != reflect.Func {
			continue
		}
		if funcField.Type.NumOut() > 1 {
			continue
		}

		var tagPath = funcField.Tag.Get("path")
		var tagArg = funcField.Tag.Get("arg")
		var tagArgs []string
		if tagArg != "" {
			tagArgs = strings.Split(tagArg, ",")
		} else {
			tagArgs = []string{}
		}

		if len(tagArgs) != funcField.Type.NumIn() {
			panic("[easy_mvc] " + argType.String() + "." + funcField.Name + "() args.len(" + fmt.Sprint(funcField.Type.NumIn()) + ") != tag arg.len(" + fmt.Sprint(len(tagArgs)) + ")!")
		}

		var httpFunc = func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			var args = []reflect.Value{}
			for i := 0; i < funcField.Type.NumIn(); i++ {
				var argItemType = funcField.Type.In(i)

				var httpArg = r.Form.Get(tagArgs[i]) //http arg
				switch argItemType.Kind() {
				case reflect.String:
					args = append(args, reflect.ValueOf(httpArg))
					break
				}
			}
			var results = field.Call(args)
			//json
			var b, _ = json.Marshal(results[0].Interface())
			w.Write(b)
		}
		http.HandleFunc(tagPath, httpFunc)
	}

}
