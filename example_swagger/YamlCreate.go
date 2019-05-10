package example_swagger

import (
	"gopkg.in/yaml.v2"
	"log"
	"reflect"
	"strings"
)

type Param struct {
	Name        string `yaml:"name"`
	In          string `yaml:"in"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

func Scan(arg interface{}) []SwaggerApi {
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
		//反射方法类型
		var funSplits = [][]string{}
		for i := 0; i < funcField.Type.NumIn(); i++ {
			var funcType = funcField.Type.In(i)
			var defs = strings.Split(tagArgs[i], ":")
			funSplits = append(funSplits, defs)

			api.Param = append(api.Param, Param{
				Name:        tagArgs[i],
				In:          "query",
				Description: noteMap[tagArgs[i]],
				Type:        funcType.Name(),
			})

		}
		api.Path = tagPath
		api.Method = "post"
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

type SwaggerApi struct {
	Param           []Param
	Controller      string
	Api_description string
	Path            string
	Method          string
}

func CreateSwaggerYaml(arg []SwaggerApi) []byte {
	root := make(map[interface{}]interface{})
	var paths = map[interface{}]interface{}{}
	for _, item := range arg {
		var paramter = []Param{}
		for _, argItem := range item.Param {
			paramter = append(paramter, Param{Name: argItem.Name, Type: argItem.Type, In: argItem.In, Description: argItem.Description})
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
