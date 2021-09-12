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
