package easy_swagger

//附加字段
type AppendParam struct {
	MustKey []SwaggerParam
	Path string  //为 "" 则全部使用 MustKey， 否则 填写具体路径为 具体接口例如 /user  匹配所有/user**
}
