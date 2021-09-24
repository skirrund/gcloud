# gcloud  
//go:embed resources/bootstrap.properties  
var baseConfig []byte  
func main() {  
	//b, _ := Asset("resources/bootstrap.properties")  
	application := bootstrap.BootstrapAll(bytes.NewReader(baseConfig), "properties")  
	defer func() {  
		if err := recover(); err != nil {  
			logger.Error("[Main] recover :", err)  
			os.Exit(0)
		}  
	}()  
	cfg := env.GetInstance()  

	application.StartMq(cfg.GetString(mq.SERVER_URL_KEY), cfg.GetInt64(mq.CONNECTION_TIMEOUT_KEY), cfg.GetInt64(mq.OPERATION_TIMEOUT_KEY))

	consumer.InitConsumer()

	application.StartWebServerWith(application.ServerOptions, func(engine *gin.Engine) {
		initRouter(engine)
	})
  
  func initRouter(router *gin.Engine) {  
	china.NewChinaIntance().InitRouter(router)  
}

# swagger
## github
	https://github.com/swaggo/gin-swagger
## 生成文档，具体参数说明见官网
	swag init --parseDependency --parseInternal
## 访问
	http://host:port/swagger/index.htm
## 注意事项
1. 需要在项目main.go中添加形如 _ "mth-maindata-service/docs" 的import
2. 
   
