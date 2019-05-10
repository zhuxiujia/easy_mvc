package easy_mvc

import (
	"github.com/facebookgo/inject"
	"sync"
)

var Context inject.Graph

func Provide(name string, value interface{}) {
	Context.Provide(&inject.Object{
		Name:  name,
		Value: value,
	})
}

var ControllerTable = sync.Map{}
