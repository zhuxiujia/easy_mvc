GO mvc框架，支持IOC,AOP,DI 基于golang标准库,http库和Tag机制实现.免生成中间代码，非常容易使用

##核心特性
* 轻量 完全兼容标准库的http，可以混合使用，扩展性高
* 过滤链支持
* 错误处理器链支持
* 使用tag 定义 http请求参数，包含 *int,*string,*float 同时支持标准库的 writer http.ResponseWriter, request *http.Request
* 支持参数默认值 只需在tag中 定义，例如 func(phone string, pwd string, age *int) interface{} arg:"phone,pwd,age:1"  其中 arg没有传参则默认为1
* 指针参数可为空（nil）非指针参数 如果没有值框架会拦截
* root path支持，类似spring controller定义一个基础的path加控制器具体方法的http path
* 内置使用tag定义文档,免除繁琐的文档编写（即将到来）