package easy_swagger

import (
	"fmt"
	"github.com/zhuxiujia/easy_mvc"
	"gopkg.in/yaml.v2"
	"log"
	"reflect"
	"strings"
)

type SwaggerParam struct {
	Name        string `yaml:"name"`
	In          string `yaml:"in"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Default     string `yaml:"default"`
	Required    bool   `yaml:"required"`
}

type SwaggerApi struct {
	Param           []SwaggerParam
	Controller      string
	Api_description string
	Path            string
	Method          string
}

//扫描上下文生成swagger的yaml
func ScanControllerContext(config SwaggerConfig) []byte {
	var swaApis = []SwaggerApi{}
	easy_mvc.ControllerTable.Range(func(key, value interface{}) bool {
		var items = Scan(value, config)
		swaApis = append(swaApis, items...)
		return true
	})
	return CreateSwaggerYaml(swaApis)
}

//扫描一个controller结构体
func Scan(arg interface{}, config SwaggerConfig) []SwaggerApi {
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
	var result = []SwaggerApi{}

	var argType = argV.Type()
	var rootPath = checkHaveRootPath(argType)
	for i := 0; i < argType.NumField(); i++ {
		var api = SwaggerApi{}
		api.Controller = argType.Name()
		var funcField = argType.Field(i)
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

		var docArg = funcField.Tag.Get("doc_arg")
		var noteMap = map[string]string{}
		var notes = strings.Split(docArg, ",")
		for _, noteItem := range notes {
			var sp = strings.Split(noteItem, ":")
			if len(sp) == 2 {
				noteMap[sp[0]] = sp[1]
			} else {
				if noteItem != "" {
					panic("[easy_mvc] note \"" + noteItem + "\"must have ':' and  value!")
				}
			}
		}

		var MustKeys []SwaggerParam
		if config.AppendParam != nil {
			for _, v := range config.AppendParam {
				if v.Path == "" {
					MustKeys = v.MustKey
				} else {
					if strings.Contains(tagPath, v.Path) {
						MustKeys = v.MustKey
					}
				}
			}
		}
		var MustKeysLen = 0
		if MustKeys != nil {
			MustKeysLen = len(MustKeys)
		}

		//反射方法类型
		var funSplits = [][]string{}
		for i := 0; i < funcField.Type.NumIn()+MustKeysLen; i++ {

			if (i + 1) > funcField.Type.NumIn() {
				api.Param = append(api.Param, MustKeys[(i - funcField.Type.NumIn())])
				continue
			}

			var funcType = funcField.Type.In(i)
			var defs = strings.Split(tagArgs[i], ":")
			funSplits = append(funSplits, defs)

			//defs[1] 为默认值
			var swaggerParam = SwaggerParam{
				Name:        defs[0],
				In:          "query",
				Description: noteMap[tagArgs[i]],
				Type:        funcType.Name(),
			}
			if len(defs) > 1 {
				swaggerParam.Default = defs[1]
			}
			if funcType.Kind() == reflect.Ptr {
				swaggerParam.Required = false
			} else {
				swaggerParam.Required = true
			}
			api.Param = append(api.Param, swaggerParam)
		}
		api.Path = tagPath
		api.Method = funcField.Tag.Get("method")
		if api.Method == "" {
			api.Method = "get"
		}
		api.Api_description = funcField.Tag.Get("doc")
		result = append(result, api)
	}
	return result
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

func CreateSwaggerYaml(arg []SwaggerApi) []byte {
	root := make(map[interface{}]interface{})
	var paths = map[interface{}]interface{}{}
	for _, item := range arg {
		var paramter = []SwaggerParam{}
		for _, argItem := range item.Param {
			switch argItem.Type {
			case "int":
				argItem.Type = "integer"
				break
			case "int16":
				argItem.Type = "integer"
				break
			case "int32":
				argItem.Type = "integer"
				break
			case "int64":
				argItem.Type = "integer"
				break
			}
			paramter = append(paramter, argItem)
		}
		var parameters = map[interface{}]interface{}{}
		parameters["tags"] = []string{item.Controller}
		parameters["parameters"] = paramter
		parameters["summary"] = item.Api_description
		parameters["description"] = item.Api_description
		parameters["responses"] = map[interface{}]interface{}{200: map[interface{}]interface{}{"description": "OK"}} //不变
		var pet = map[interface{}]interface{}{}
		pet[item.Method] = parameters
		paths[item.Path] = pet
	}

	root["paths"] = paths
	root["swagger"] = "2.0"
	root["info"] = map[interface{}]interface{}{
		"version":        "",
		"title":          "",
		"description":    "",
		"termsOfService": "",
	}
	var controllers = []map[interface{}]interface{}{}
	for _, item := range arg {
		controllers = append(controllers, map[interface{}]interface{}{"name": item.Controller, "description": ""})
	}
	root["tags"] = controllers
	d, _ := yaml.Marshal(&root)
	return d
}
