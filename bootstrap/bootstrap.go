package bootstrap

import (
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/utils/idworker"

	"github.com/skirrund/gcloud/cache/redis"
	"github.com/skirrund/gcloud/config"
	"github.com/skirrund/gcloud/logger"

	"github.com/skirrund/gcloud/registry"

	db "github.com/skirrund/gcloud/datasource"
	"github.com/skirrund/gcloud/mq"

	gfiber "github.com/skirrund/gcloud/plugins/server/http/gfiber"

	"github.com/skirrund/gcloud/bootstrap/env"
)

type Options struct {
	Registry     registry.IRegistry
	Mq           mq.IClient
	Redis        *redis.RedisClient
	IdWorker     *idworker.Worker
	ConfigCenter config.IConfig
}

type Application struct {
	BootOptions   BootstrapOptions
	ServerOptions server.Options
	Registry      registry.IRegistry
	Mq            mq.IClient
	Redis         *redis.RedisClient
	IdWorker      *idworker.Worker
	ConfigCenter  config.IConfig
}

type BootstrapOptions struct {
	Profile       string
	ServerAddress string
	ServerPort    uint64
	ServerName    string
	Host          string
	LoggerDir     string
	LoggerConsole bool
	LoggerJson    bool
	Config        config.IConfig
}

var MthApplication *Application

func StartBase(reader io.Reader, fileType string) *Application {
	bo := initBaseOptions(reader, fileType)
	if MthApplication != nil {
		MthApplication.BootOptions = bo
	} else {
		MthApplication = &Application{
			BootOptions: bo,
		}

		log.Println("[Bootstrap]init BootstrapOptions properties:[Profile=" + bo.Profile + "]" + ",[ServerName=" + bo.ServerName + "],[Bind=" + bo.ServerAddress + "]" + ",[LoggerDir=" + bo.LoggerDir + "]")
	}
	so := server.Options{
		ServerName: bo.ServerName,
		Address:    bo.ServerAddress,
	}
	MthApplication.ServerOptions = so
	return MthApplication
}

func initBaseOptions(reader io.Reader, fileType string) BootstrapOptions {
	cfg := env.GetInstance()
	cfg.SetBaseConfig(reader, fileType)
	host, _ := os.Hostname()
	address := cfg.GetString(env.SERVER_ADDRESS_KEY)
	port := uint64(8080)
	profile := cfg.GetString(env.SERVER_PROFILE_KEY)
	sn := cfg.GetString(env.SERVER_SERVERNAME_KEY)
	ld := cfg.GetString(env.LOGGER_DIR_KEY)
	cfgFile := cfg.GetString(env.SERVER_CONFIGFILE_KEY)
	var flagProfile string
	var flagCfgFile string
	var flagSn string
	var flagAddress string
	var flagLogdir string
	var flagLogMaxAge uint64
	var flagConsoleLog bool
	var flagJsonLog bool
	flag.StringVar(&flagProfile, env.SERVER_PROFILE_KEY, "", "server profile:[dev,test,prod...]")
	flag.StringVar(&flagCfgFile, env.SERVER_CONFIGFILE_KEY, "", "server config file")
	flag.StringVar(&flagSn, env.SERVER_SERVERNAME_KEY, "", "server name")
	flag.StringVar(&flagAddress, env.SERVER_ADDRESS_KEY, "", "server address")
	flag.StringVar(&flagLogdir, env.LOGGER_DIR_KEY, "", "logDir")
	flag.Uint64Var(&flagLogMaxAge, env.LOGGER_MAXAGE_KEY, 7, "log maxAge:day   default:7")
	flag.BoolVar(&flagConsoleLog, env.LOGGER_CONSOLE, true, "logger.console enabled:{default:true}")
	flag.BoolVar(&flagJsonLog, env.LOGGER_JSON, false, "logger.json enabled:{default:false}")
	flag.Parse()
	if len(flagProfile) == 0 {
		flagProfile = profile
	}
	if len(flagCfgFile) == 0 {
		flagCfgFile = cfgFile
	}
	if len(flagSn) == 0 {
		flagSn = sn
	}
	if len(flagAddress) == 0 {
		flagAddress = address
	}
	var err error
	if len(flagAddress) > 0 {
		port, err = strconv.ParseUint(strings.Split(flagAddress, ":")[1], 10, 64)
		if err != nil {
			panic(err)
		}
	} else {
		panic("server.address error[" + flagAddress + "]")
	}

	if len(flagLogdir) == 0 {
		flagLogdir = ld
	}
	cfg.Set(env.SERVER_ADDRESS_KEY, flagAddress)
	cfg.Set(env.SERVER_PORT_KEY, port)
	cfg.Set(env.SERVER_PROFILE_KEY, flagProfile)
	cfg.Set(env.SERVER_SERVERNAME_KEY, flagSn)
	cfg.Set(env.LOGGER_DIR_KEY, flagLogdir)
	cfg.Set(env.LOGGER_MAXAGE_KEY, flagLogMaxAge)
	cfg.Set(env.LOGGER_CONSOLE, flagConsoleLog)
	cfg.Set(env.SERVER_CONFIGFILE_KEY, flagCfgFile)
	cfg.Set(env.LOGGER_JSON, flagJsonLog)
	cfg.LoadProfileBaseConfig(flagProfile, fileType)
	return BootstrapOptions{
		ServerAddress: flagAddress,
		ServerPort:    port,
		Profile:       flagProfile,
		ServerName:    flagSn,
		LoggerDir:     flagLogdir,
		LoggerConsole: flagConsoleLog,
		LoggerJson:    flagJsonLog,
		Host:          host,
		Config:        env.GetInstance(),
	}
}

func (app *Application) Bootstrap(options Options) {
	app.StartLogger()
	app.ConfigCenter = options.ConfigCenter
	app.Registry = options.Registry
	app.Redis = options.Redis
	if options.Redis != nil {
		worker, _ := idworker.NewWorkerWithRedis(options.Redis, MthApplication.BootOptions.ServerName)
		app.IdWorker = worker
	}
	app.Mq = options.Mq
}

func (app *Application) BootstrapAll(options Options) {
	app.Bootstrap(options)
	//MthApplication.SentinelInit()
	//MthApplication.StartRedis()
	MthApplication.StartDb()
}

func (app *Application) StartLogger() {
	ops := app.BootOptions
	maxAge := env.GetInstance().GetUint64WithDefault(env.LOGGER_MAXAGE_KEY, 7)
	logger.InitLog(ops.LoggerDir, ops.ServerName, strconv.FormatUint(ops.ServerPort, 10), ops.LoggerConsole, ops.LoggerJson, maxAge)
}

func (app *Application) StartDb() {
	cfg := env.GetInstance()
	db.InitDataSource(db.Option{
		DSN:             cfg.GetString(db.DB_DSN),
		MaxIdleConns:    cfg.GetInt(db.DB_MAX_IDLE_CONNS),
		MaxOpenConns:    cfg.GetInt(db.DB_MAX_OPEN_CONNS),
		ConnMaxLifetime: cfg.GetInt(db.DB_CONN_MAX_LIFE_TIME),
	})
}

func (app *Application) ShutDown() {

	if cfg := app.ConfigCenter; cfg != nil {
		cfg.Shutdown()
	}
	if registry := app.Registry; registry != nil {
		registry.Shutdown()
	}
	if mq := app.Mq; mq != nil {
		mq.Close()
	}
	if redisClient := app.Redis; redisClient != nil {
		redisClient.Close()
	}
	logger.Sync()
}

func (app *Application) StartWebServerWith(options server.Options, routerProvider func(engine *fiber.App), middleware ...fiber.Handler) {
	middlewares := make([]any, len(middleware))
	for i, m := range middleware {
		middlewares[i] = m
	}
	srv := gfiber.NewServer(options, routerProvider, middlewares...)
	if app.Registry != nil {
		delayFunction(func() {
			err := app.Registry.RegisterInstance()
			if err != nil {
				logger.Panic("[Bootstrap] registerInstance error", err.Error())
			}
		})
	}
	srv.Run(app.ShutDown)
}

func (app *Application) StartWebServer(routerProvider func(engine *fiber.App), middleware ...fiber.Handler) {
	ops := app.BootOptions
	options := server.Options{
		ServerName: ops.ServerName,
		Address:    ops.ServerAddress,
	}
	app.StartWebServerWith(options, routerProvider, middleware...)
}

func delayFunction(f func()) {
	time.AfterFunc(1*time.Second, func() {
		f()
	})
}

// func sentinelNacosInit(entity *sentinel_config.Entity) bool {
// 	//nacos server地址
// 	serverAddrStr := env.GetInstance().GetString("sentinel.datasource.nacos.server-addr")
// 	if len(serverAddrStr) == 0 {
// 		return false
// 	}
// 	var scs []constant.ServerConfig
// 	serverAddrs := strings.Split(serverAddrStr, ",")
// 	for _, serverAddr := range serverAddrs {
// 		urlAndPort := strings.Split(serverAddr, ":")
// 		port := 8848
// 		if len(urlAndPort) > 1 {
// 			var err error
// 			port, err = strconv.Atoi(urlAndPort[1])
// 			if err != nil {
// 				port = 8848
// 			}
// 		}
// 		sc := constant.ServerConfig{
// 			ContextPath: "/nacos",
// 			Port:        uint64(port),
// 			IpAddr:      urlAndPort[0],
// 		}
// 		scs = append(scs, sc)
// 	}

// 	//nacos client 相关参数配置,具体配置可参考https://github.com/nacos-group/nacos-sdk-go
// 	// https://sentinelguard.io/zh-cn/docs/golang/hotspot-param-flow-control.html
// 	cc := constant.ClientConfig{
// 		TimeoutMs:   5000,
// 		NamespaceId: env.GetInstance().GetString("sentinel.datasource.nacos.namespace"),
// 		LogDir:      entity.LogBaseDir() + "/nacos",
// 	}
// 	//生成nacos config client(配置中心客户端)
// 	client, err := clients.CreateConfigClient(map[string]interface{}{
// 		"serverConfigs": scs,
// 		"clientConfig":  cc,
// 	})
// 	if err != nil {
// 		logger.Errorf("Fail to create client, err: %+v", err)
// 		return false
// 	}
// 	// 注册流控规则Handler
// 	h := sentinel_ds.NewFlowRulesHandler(sentinel_ds.FlowRuleJsonArrayParser)
// 	registerAndInitDs(client, h, "-flow-rule")
// 	// 注册热点规则
// 	h2 := sentinel_ds.NewHotSpotParamRulesHandler(sentinel_ds.HotSpotParamRuleJsonArrayParser)
// 	registerAndInitDs(client, h2, "-param-flow-rule")
// 	// 注册溶断规则
// 	h3 := sentinel_ds.NewCircuitBreakerRulesHandler(sentinel_ds.CircuitBreakerRuleJsonArrayParser)
// 	registerAndInitDs(client, h3, "-degrade-rule")
// 	// 系统规则
// 	h4 := sentinel_ds.NewSystemRulesHandler(sentinel_ds.SystemRuleJsonArrayParser)
// 	registerAndInitDs(client, h4, "-system-rule")
// 	// 授权规则
// 	h5 := sentinel_ds.NewIsolationRulesHandler(sentinel_ds.IsolationRuleJsonArrayParser)
// 	registerAndInitDs(client, h5, "-authority-rule")
// 	return true
// }

// func registerAndInitDs(client config_client.IConfigClient, h sentinel_ds.PropertyHandler, dataIdSuffix string) {
// 	//创建NacosDataSource数据源
// 	//sentinel-go 对应在nacos中创建配置文件的group
// 	//flow 对应在nacos中创建配置文件的dataId
// 	nds, err := sentinel_nacos.NewNacosDataSource(client, env.GetInstance().GetString("sentinel.datasource.nacos.groupId"),
// 		env.GetInstance().GetString("server.name")+dataIdSuffix, h)
// 	if err != nil {
// 		logger.Errorf("Fail to create nacos data source client, err: %+v", err)
// 		return
// 	}
// 	//nacos数据源初始化
// 	err = nds.Initialize()
// 	if err != nil {
// 		logger.Errorf("Fail to initialize nacos data source client, err: %+v", err)
// 		return
// 	}
// }

// func sentinelConfigInit() (*sentinel_config.Entity, error) {
// 	entity := sentinel_config.NewDefaultConfig()
// 	// 设置相关默认值
// 	entity.Sentinel.Log.Metric.MaxFileCount = 14
// 	// 100MB
// 	entity.Sentinel.Log.Metric.SingleFileMaxSize = 104857600
// 	ParseSentinelConfig(entity, "resources/sentinel.yaml")
// 	if entity.Sentinel.App.Name == sentinel_config.UnknownProjectName {
// 		entity.Sentinel.App.Name = env.GetInstance().GetString("server.name")
// 	}
// 	if entity.Sentinel.Log.Dir == sentinel_config.GetDefaultLogDir() {
// 		entity.Sentinel.Log.Dir = env.GetInstance().GetString("logger.dir") + "/" + entity.Sentinel.App.Name + "/csp/"
// 	}
// 	err := sentinel.InitWithConfig(entity)
// 	if err != nil {
// 		logger.Errorf("sentinel config init error, %v", err.Error())
// 		return entity, err
// 	}
// 	return entity, nil
// }

// func ParseSentinelConfig(entity *sentinel_config.Entity, filePath string) error {
// 	_, err := os.Stat(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	content, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		logger.Errorf("sentinel config read sentinel.yaml error," + err.Error())
// 		return err
// 	}
// 	err = yaml.Unmarshal(content, entity)
// 	if err != nil {
// 		logger.Errorf("sentinel config Unmarshal error, %v", err.Error())
// 		return err
// 	}
// 	logging.Info("[Config] Resolving Sentinel config from file=" + filePath + " success")
// 	return nil
// }

// // 初始化Sentinel
// func (app *Application) SentinelInit() {
// 	if env.GetInstance().GetString("spring.cloud.sentinel.enabled") == "false" {
// 		return
// 	}
// 	entity, err := sentinelConfigInit()
// 	if err == nil && entity != nil {
// 		sentinelNacosInit(entity)
// 	}
// }
