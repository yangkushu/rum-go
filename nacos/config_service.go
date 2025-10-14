package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type ConfigService struct {
	configClient config_client.IConfigClient
	config       *Config
}

func NewConfigService(config *Config) (*ConfigService, error) {
	clientConfig := constant.ClientConfig{
		//Namespace:         "e525eafa-f7d7-4029-83d9-008937f9d468", // 如果需要支持多namespace，我们可以创建多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		//TimeoutMs:           5000,
		//NotLoadCacheAtStart: true,
		//LogDir:              "/tmp/nacos/log",
		//CacheDir:            "/tmp/nacos/cache",
		//LogLevel:            "debug"
		TimeoutMs:    config.Timeout,      // 请求Nacos服务端的超时时间，默认是10000m
		BeatInterval: config.BeatInterval, // 向服务器发送心跳间隔时间，单位毫秒，默认 5000ms
		NamespaceId:  config.Namespace,    // nacos命名空间
		//Endpoint:             "",                                      // 获取nacos节点ip的服务地址
		//CacheDir:             "data/cache",                            // 缓存目录
		//LogDir:               "data/log",                              // 日志目录
		CacheDir:             config.CacheDir,        // 缓存目录
		LogDir:               config.LogDir,          // 日志目录
		UpdateThreadNum:      config.UpdateThreadNum, // 更新服务信息使用的gorutine线程数，默认20
		NotLoadCacheAtStart:  true,                   // 在启动时不读取本地缓存数据，true--不读取，false--读取
		UpdateCacheWhenEmpty: true,                   // 当服务列表为空时是否更新本地缓存，true--更新,false--不更新
		LogLevel:             config.LogLevel,
	}

	if len(config.Username) > 0 && len(config.Password) > 0 {
		clientConfig.Username = config.Username
		clientConfig.Password = config.Password
	}

	// 服务端的配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      config.Addr, // nacos服务的ip地址
			Port:        config.Port, // nacos服务端口n
			ContextPath: config.Path, // nacos服务的上下文路径，默认是“/nacos”
			Scheme:      config.Scheme,
		},
	}

	// 创建动态配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, err
	}

	return &ConfigService{configClient: configClient, config: config}, nil
}

func (c *ConfigService) GetConfig() (string, error) {
	return c.configClient.GetConfig(vo.ConfigParam{
		DataId: c.config.DataId,
		Group:  c.config.Group,
	})
}

func (c *ConfigService) Close() error {
	return nil
}
