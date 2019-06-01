package easy_swagger

import (
	"github.com/rakyll/statik/fs"
	_ "github.com/zhuxiujia/easy_mvc/easy_swagger/dist/statik"
	"log"
	"net/http"
	"strings"
)

const htmlTemplete = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="./swagger_static/swagger-ui.css" >
    <link rel="icon" type="image/png" href="./swagger_static/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./swagger_static/favicon-16x16.png" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }

      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
    <script src="/swagger_static/swagger-ui-bundle.js"></script>
  </head>

  <body>
    <div id="swagger-ui"></div>

    <script src="./swagger_static/swagger-ui-bundle.js"> </script>
    <script src="./swagger_static/swagger-ui-standalone-preset.js"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        url: window.location.protocol+'//'+window.location.host+"#{serverAddr}",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region

      window.ui = ui
    }
  </script>
  </body>
</html>
`

type IndexHtmlHandle struct {
	html string
}

func (it IndexHtmlHandle) New(addr string) IndexHtmlHandle {
	it.html = strings.Replace(htmlTemplete, "#{serverAddr}", addr, -1)
	return it
}

func (it *IndexHtmlHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(it.html))
}

//启用swagger ui 配置
func EnableSwagger(serverAddr string,config SwaggerConfig) {
	var bytes = ScanControllerContext(config) //扫描上下文环境
	//doc接口放出swagger yaml配置
	http.HandleFunc("/doc", func(w http.ResponseWriter, r *http.Request) {
		w.Write(bytes)
	})
	statikFS, _ := fs.New()

	//swagger ui 必须的js文件
	http.Handle("/swagger_static/", http.FileServer(statikFS))

	if !strings.Contains(serverAddr, "http://") {
		serverAddr = "http://" + serverAddr
	}
	var h = IndexHtmlHandle{}.New(  "/doc")
	http.Handle("/swagger", &h)

	log.Println("[easy_mvc] swagger ui yaml config on :" + serverAddr + "/doc")
	log.Println("[easy_mvc] swagger ui web start on :" + serverAddr + "/swagger")
	//http.ListenAndServe(serverAddr, nil) 这里由用户构建（最后调用）
}
