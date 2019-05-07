package easy_mvc

import "github.com/facebookgo/inject"

var Context inject.Graph

func Provide(name string, value interface{}) {
	Context.Provide(&inject.Object{
		Name:  name,
		Value: value,
	})
}
