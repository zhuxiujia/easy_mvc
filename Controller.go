package easy_mvc

import "net/http"

type ActivityController struct {
	Add2 func(writer http.ResponseWriter, request *http.Request)
	Add3 func(id string, name string) (string, error) `path:"/admin/activity/list" args:"id,name" method:"post"`
}
