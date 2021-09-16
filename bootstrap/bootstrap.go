package bootstrap

import (
	"flag"
	sentinel "github.com/alibaba/sentinel-golang/api"
	sentinel_ds "github.com/alibaba/sentinel-golang/ext/datasource"
	sentinel_config "github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	sentinel_nacos "github.com/skirrund/gcloud/plugins/sentinel"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/skirrund/gcloud/server"

	"github.com/skirrund/gcloud/cache/redis"
	"github.com/skirrund/gcloud/config"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils"

	nacosConfig "github.com/skirrund/gcloud/plugins/nacos/config"
	nacosRegistry "github.com/skirrund/gcloud/plugins/nacos/registry"
	"github.com/skirrund/gcloud/registry"

	"github.com/skirrund/gcloud/datasource"
	"github.com/skirrund/gcloud/mq"

	mthGin "github.com/skirrund/gcloud/plugins/server/http/gin"

	"github.com/skirrund/gcloud/bootstrap/env"

	"github.com/gin-gonic/gin"
)

type Application struct {
	BootOptions   BootstrapOptions
	ServerOptions server.Options
	Registry      registry.IRegistry
	Mq            *mq.Client
	Redis         *redis.RedisClient
	IdWorker      *utils.Worker
	ConfigCenter  config.IConfig
}

type BootstrapOptions struct {
	Profile       string
	ServerAddress string
	ServerPort    uint64
	ServerName    string
	Host          string
	LoggerDir     string
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
	var flagProfile string
	var flagSn string
	var flagAddress string
	var flagLogdir string
	var flagLogMaxAge uint64
	flag.StringVar(&flagProfile, env.SERVER_PROFILE_KEY, "", "server profile:[dev,local,prod]")
	flag.StringVar(&flagSn, env.SERVER_SERVERNAME_KEY, "", "sererver name")
	flag.StringVar(&flagAddress, env.SERVER_ADDRESS_KEY, "", "sererver address")
	flag.StringVar(&flagLogdir, env.LOGGER_DIR_KEY, "", "logDir")
	flag.Uint64Var(&flagLogMaxAge, env.LOGGER_MAXAGE_KEY, 7, "log maxAge:day   default:7")
	flag.Parse()
	if len(flagProfile) == 0 {
		flagProfile = profile
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

	return BootstrapOptions{
		ServerAddress: flagAddress,
		ServerPort:    port,
		Profile:       flagProfile,
		ServerName:    flagSn,
		LoggerDir:     flagLogdir,
		Host:          host,
		Config:        env.GetInstance(),
	}
}

func BootstrapAll(reader io.Reader, fileType string) *Application {
	if MthApplication == nil {
		MthApplication = StartBase(reader, fileType)
	}
	MthApplication.StartLogger()
	MthApplication.StartConfigCenter()
	MthApplication.SentinelInit()
	MthApplication.StartRegistry()
	MthApplication.StartRedis()
	MthApplication.StartDb()
	return MthApplication
}

func (app *Application) StartLogger() {
	ops := app.BootOptions
	maxAge := env.GetInstance().GetUint64WithDefault(env.LOGGER_MAXAGE_KEY, 7)
	if app.BootOptions.Profile == "local" {
		logger.InitLog(ops.LoggerDir, ops.ServerName, strconv.FormatUint(ops.ServerPort, 10), true, maxAge)
	} else {
		logger.InitLog(ops.LoggerDir, ops.ServerName, strconv.FormatUint(ops.ServerPort, 10), false, maxAge)
	}
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

func (app *Application) StartConfigCenter() {
	cfg := env.GetInstance()
	addr := cfg.GetString(nacosConfig.NACOS_CONFIG_SERVER_ADDR_KEY)
	fe := cfg.GetString(nacosConfig.NACOS_CONFIG_FILE_EXTENSION_KEY)
	group := cfg.GetString(nacosConfig.NACOS_CONFIG_GROUP_KEY)
	ns := cfg.GetString(nacosConfig.NACOS_CONFIG_NAMESPACE_KEY)
	prefix := cfg.GetString(nacosConfig.NACOS_CONFIG_PREFIX_KEY)
	dir := app.BootOptions.LoggerDir + "/" + app.BootOptions.ServerName + "/" + app.BootOptions.Host
	logger.Info("[Bootstrap] start init nacos config center properties:[addrs=" + addr + "]" + ",[FileExtension=" + fe + "],[Group=" + group + "],[Prefix=" + prefix + "],[Namespace=" + ns + "],[Env=" + app.BootOptions.Profile + "]")

	options := config.Options{
		ServerAddrs: strings.Split(addr, ","),
		ClientOptions: config.ClientOptions{
			NamespaceId: ns,
			LogDir:      dir,
			//CacheDir:    dir,
			TimeoutMs: 5000,
			AppName:   app.BootOptions.ServerName,
		},
		ConfigOptions: config.ConfigOptions{
			Prefix:        prefix,
			FileExtension: fe,
			Env:           app.BootOptions.Profile,
			Group:         group,
		},
	}
	nc := nacosConfig.CreateInstance(options)
	app.ConfigCenter = nc
	//app.Config = config.ConfigServer
	//return config.ConfigServer
}

func (app *Application) StartRedis() *redis.RedisClient {
	opts := redis.Options{}
	utils.NewOptions(env.GetInstance(), &opts)
	client := redis.NewClient(opts)
	worker, _ := utils.NewWorkerWithRedis(client, app.BootOptions.ServerName)
	app.IdWorker = worker
	app.Redis = client
	return client
}

func (app *Application) StartMq(serverUrl string, connectionTimeOut int64, operationTimeOut int64) *mq.Client {
	client := mq.InitClient(serverUrl, connectionTimeOut, operationTimeOut, app.BootOptions.ServerName)
	app.Mq = client
	return client
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

func (app *Application) StartWebServerWith(options server.Options, routerProvider func(engine *gin.Engine)) {
	srv := mthGin.NewServer(options, routerProvider)
	srv.Run(app.ShutDown)
}

func (app *Application) StartWebServer(routerProvider func(engine *gin.Engine)) {
	ops := app.BootOptions
	options := server.Options{
		ServerName: ops.ServerName,
		Address:    ops.ServerAddress,
	}
	app.StartWebServerWith(options, routerProvider)
}

func (app *Application) StartRegistry() registry.IRegistry {
	addr := app.BootOptions.Config.GetString(nacosRegistry.NACOS_DISCOVERY_SERVER_ADDE_KEY)
	dir := app.BootOptions.LoggerDir + "/" + app.BootOptions.ServerName + "/" + app.BootOptions.Host

	options := registry.Options{
		ServerAddrs: strings.Split(addr, ","),
		ClientOptions: registry.ClientOptions{
			//NamespaceId: ns,
			LogDir: dir,
			//CacheDir:  dir,
			TimeoutMs: 3000,
			AppName:   app.BootOptions.ServerName,
		},
		RegistryOptions: registry.RegistryOptions{
			ServiceName: app.BootOptions.ServerName,
			ServicePort: app.BootOptions.ServerPort,
		},
	}
	registry.RegistryCenter = nacosRegistry.NewRegistry(options)
	delayFunction(func() {
		err := registry.RegistryCenter.RegisterInstance()
		if err != nil {
			logger.Panic("[Bootstrap] registerInstance error", err.Error())
		}
	})
	app.Registry = registry.RegistryCenter
	return registry.RegistryCenter
}

func delayFunction(f func()) {
	timer := time.NewTimer(1 * time.Second)

	select {
	case <-timer.C:
		f()
	}
}

func sentinelNacosInit(entity *sentinel_config.Entity) bool {
	//nacos server地址
	serverAddrStr := env.GetInstance().GetString("sentinel.datasource.nacos.server-addr")
	if len(serverAddrStr) == 0 {
		return false
	}
	var scs []constant.ServerConfig
	serverAddrs := strings.Split(serverAddrStr, ",")
	for _, serverAddr := range serverAddrs {
		urlAndPort := strings.Split(serverAddr, ":")
		port := 8848
		if len(urlAndPort) > 1 {
			var err error
			port, err = strconv.Atoi(urlAndPort[1])
			if err != nil {
				port = 8848
			}
		}
		sc := constant.ServerConfig{
			ContextPath: "/nacos",
			Port:        uint64(port),
			IpAddr:      urlAndPort[0],
		}
		scs = append(scs, sc)
	}

	//nacos client 相关参数配置,具体配置可参考https://github.com/nacos-group/nacos-sdk-go
	// https://sentinelguard.io/zh-cn/docs/golang/hotspot-param-flow-control.html
	cc := constant.ClientConfig{
		TimeoutMs: 5000,
		NamespaceId: env.GetInstance().GetString("sentinel.datasource.nacos.namespace"),
		LogDir: entity.LogBaseDir() + "/nacos",
	}
	//生成nacos config client(配置中心客户端)
	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": scs,
		"clientConfig":  cc,
	})
	if err != nil {
		logger.Errorf("Fail to create client, err: %+v", err)
		return false
	}
	// 注册流控规则Handler
	h := sentinel_ds.NewFlowRulesHandler(sentinel_ds.FlowRuleJsonArrayParser)
	registerAndInitDs(client, h, "-flow-rule")
	// 注册热点规则
	h2 := sentinel_ds.NewHotSpotParamRulesHandler(sentinel_ds.HotSpotParamRuleJsonArrayParser)
	registerAndInitDs(client, h2, "-param-flow-rule")
	// 注册溶断规则
	h3 := sentinel_ds.NewCircuitBreakerRulesHandler(sentinel_ds.CircuitBreakerRuleJsonArrayParser)
	registerAndInitDs(client, h3, "-degrade-rule")
	// 系统规则
	h4 := sentinel_ds.NewSystemRulesHandler(sentinel_ds.SystemRuleJsonArrayParser)
	registerAndInitDs(client, h4, "-system-rule")
	// 授权规则
	h5 := sentinel_ds.NewIsolationRulesHandler(sentinel_ds.IsolationRuleJsonArrayParser)
	registerAndInitDs(client, h5, "-authority-rule")
	return true
}

func registerAndInitDs(client config_client.IConfigClient, h sentinel_ds.PropertyHandler, dataIdSuffix string)  {
	//创建NacosDataSource数据源
	//sentinel-go 对应在nacos中创建配置文件的group
	//flow 对应在nacos中创建配置文件的dataId
	nds, err := sentinel_nacos.NewNacosDataSource(client, env.GetInstance().GetString("sentinel.datasource.nacos.groupId"),
		env.GetInstance().GetString("server.name") + dataIdSuffix, h)
	if err != nil {
		logger.Errorf("Fail to create nacos data source client, err: %+v", err)
		return
	}
	//nacos数据源初始化
	err = nds.Initialize()
	if err != nil {
		logger.Errorf("Fail to initialize nacos data source client, err: %+v", err)
		return
	}
}

func sentinelConfigInit() (*sentinel_config.Entity, error) {
	entity := sentinel_config.NewDefaultConfig()
	// 设置相关默认值
	entity.Sentinel.Log.Metric.MaxFileCount = 14
	// 100MB
	entity.Sentinel.Log.Metric.SingleFileMaxSize = 104857600
	ParseSentinelConfig(entity, "resources/sentinel.yaml")
	if entity.Sentinel.App.Name == sentinel_config.UnknownProjectName {
		entity.Sentinel.App.Name = env.GetInstance().GetString("server.name")
	}
	if entity.Sentinel.Log.Dir == sentinel_config.GetDefaultLogDir() {
		entity.Sentinel.Log.Dir = env.GetInstance().GetString("logger.dir") + "/" + entity.Sentinel.App.Name + "/csp/"
	}
	err := sentinel.InitWithConfig(entity)
	if err != nil {
		logger.Errorf("sentinel config init error, %v", err.Error())
		return entity, err
	}
	return entity, nil
}

func ParseSentinelConfig(entity *sentinel_config.Entity, filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Errorf("sentinel config read sentinel.yaml error," + err.Error())
		return err
	}
	err = yaml.Unmarshal(content, entity)
	if err != nil {
		logger.Errorf("sentinel config Unmarshal error, %v", err.Error())
		return err
	}
	logging.Info("[Config] Resolving Sentinel config from file=" + filePath + " success")
	return nil
}

// 初始化Sentinel
func (app *Application) SentinelInit() {
	if env.GetInstance().GetString("spring.cloud.sentinel.enabled") == "false" {
		return
	}
	entity, err := sentinelConfigInit()
	if err == nil && entity != nil {
		sentinelNacosInit(entity)
	}
}