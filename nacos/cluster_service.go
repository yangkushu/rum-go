// nacos.go

package nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/pkg/errors"
	"github.com/yangkushu/rum-go/log"
	"gopkg.in/yaml.v3"
	"net"
	"net/url"
	"strconv"
	"strings"

	//"github.com/nacos-group/nacos-sdk-go/v2/clients"
	//"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	//"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	//"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	//"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/yangkushu/rum-go/utils"
	"os"
	"sync"
)

const (
	// 实例的默认权重
	defaultWeight float64 = 10.0
)

type InstanceInfo struct {
	InstanceId  string
	ServiceName string
	Group       string
	ClusterName string
	Ip          string
	Port        uint64
	Weight      float64
	Enable      bool
	Hostname    string
}

func getInstanceInfo(group string, info *model.SubscribeService) InstanceInfo {
	return InstanceInfo{
		InstanceId:  info.InstanceId,
		ServiceName: info.ServiceName,
		Group:       group,
		ClusterName: info.ClusterName,
		Ip:          info.Ip,
		Port:        info.Port,
		Weight:      info.Weight,
		Enable:      info.Enable,
	}
}

// ClusterService 集群服务对象
type ClusterService struct {
	nacosConfig *Config
	//instanceConfig  config.Instance
	client          naming_client.INamingClient
	config          config_client.IConfigClient
	deregisterParam vo.DeregisterInstanceParam
	registerParam   vo.RegisterInstanceParam

	selfInstanceId string
	localIp        string
	localPort      int
	hostname       string
	localAddress   string

	// 所有实例列表  servicename: []InstanceInfo
	instances sync.Map

	//监控列表  servicename: *EventHandler
	subscribedServices map[string]*EventHandler

	// 当前负载
	Weight float64

	// 互斥锁
	mutex sync.Mutex
}

func (s *ClusterService) GetLocalAddress() string {
	return s.localAddress
}

// 是否初始化完成
func (s *ClusterService) IsReady() bool {
	return nil != s.client
}

func (s *ClusterService) GetLocalIp() string {
	return s.localIp
}

func (s *ClusterService) GetLocalPort() int {
	return s.localPort
}

// 获取本进程服务名
func (s *ClusterService) GetServiceName() string {
	if nil != s.nacosConfig {
		return s.nacosConfig.Service
	} else {
		return ""
	}
}

// 获取本进程组名
func (s *ClusterService) GetGroupName() string {
	if nil != s.nacosConfig {
		return s.nacosConfig.Group
	} else {
		return ""
	}
}

// 获取本机IP地址
//func (s *ClusterService) GetLocalIp() string {
//	if 0 == len(s.localIp) {
//		s.localIp = util.LocalIP()
//	}
//	return s.localIp
//}

// GetHostname 获取本机主机名
func (s *ClusterService) GetHostname() string {
	if 0 == len(s.hostname) {
		if hostname, err := os.Hostname(); nil != err {
			s.hostname = ""
		} else {
			s.hostname = hostname
		}
	}
	return s.hostname
}

//
//func (s *ClusterService) GetSelfInstanceId() string {
//	return s.selfInstanceId
//}
//

func createNacosClient(nacosConfig *Config, password string) (naming_client.INamingClient, error) {
	cc := &constant.ClientConfig{
		TimeoutMs:    nacosConfig.Timeout,      // 请求Nacos服务端的超时时间，默认是10000m
		BeatInterval: nacosConfig.BeatInterval, // 向服务器发送心跳间隔时间，单位毫秒，默认 5000ms
		NamespaceId:  nacosConfig.Namespace,    // nacos命名空间
		//Namespace: "public", // nacos命名空间
		//Endpoint:             "",                                      // 获取nacos节点ip的服务地址
		CacheDir:             nacosConfig.CacheDir,        // 缓存目录
		LogDir:               nacosConfig.LogDir,          // 日志目录
		UpdateThreadNum:      nacosConfig.UpdateThreadNum, // 更新服务信息使用的gorutine线程数，默认20
		NotLoadCacheAtStart:  true,                        // 在启动时不读取本地缓存数据，true--不读取，false--读取
		UpdateCacheWhenEmpty: true,                        // 当服务列表为空时是否更新本地缓存，true--更新,false--不更新
		LogLevel:             nacosConfig.LogLevel,
		Username:             nacosConfig.Username,
		Password:             password,
	}

	// TODO 先只取第一个地址
	if strings.Contains(nacosConfig.Addr, ",") {
		nacosConfig.Addr = strings.Split(nacosConfig.Addr, ",")[0]
	}

	scheme := ""
	// 去掉http:// https://
	if strings.HasPrefix(nacosConfig.Addr, "http://") {
		scheme = "http"
		nacosConfig.Addr = nacosConfig.Addr[7:]
	}

	if strings.HasPrefix(nacosConfig.Addr, "https://") {
		scheme = "https"
		nacosConfig.Addr = nacosConfig.Addr[8:]
	}

	if strings.Contains(nacosConfig.Addr, ":") {
		split := strings.Split(nacosConfig.Addr, ":")
		if len(split) == 2 {
			nacosConfig.Addr = split[0]
			port, err := strconv.Atoi(split[1])
			if err != nil {
				log.Error("parse nacos addr port error", log.ErrorField(err))
				return nil, errors.Wrap(err, "parse nacos addr port error")
			}
			nacosConfig.Port = uint64(port)
		}
	}

	log.Info("nacos addr", log.String("addr", nacosConfig.Addr), log.Uint64("port", nacosConfig.Port))

	// 服务端的配置
	var serverOptions []constant.ServerOption
	serverOptions = append(serverOptions, constant.WithContextPath(nacosConfig.Path))
	if len(scheme) > 0 {
		serverOptions = append(serverOptions, constant.WithScheme(scheme))
	}
	ss := []constant.ServerConfig{
		*constant.NewServerConfig(nacosConfig.Addr, nacosConfig.Port, serverOptions...),
	}

	// 创建 naming
	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: ss,
		},
	)
}

func createNacosClientHighAvailability(nacosConfig *Config, password string) (naming_client.INamingClient, config_client.IConfigClient, error) {
	clientConfig := constant.ClientConfig{
		TimeoutMs:    nacosConfig.Timeout,      // 请求Nacos服务端的超时时间，默认是10000m
		BeatInterval: nacosConfig.BeatInterval, // 向服务器发送心跳间隔时间，单位毫秒，默认 5000ms
		NamespaceId:  nacosConfig.Namespace,    // nacos命名空间
		//Namespace: "public", // nacos命名空间
		//Endpoint:             "",                                      // 获取nacos节点ip的服务地址
		CacheDir:             nacosConfig.CacheDir,        // 缓存目录
		LogDir:               nacosConfig.LogDir,          // 日志目录
		UpdateThreadNum:      nacosConfig.UpdateThreadNum, // 更新服务信息使用的gorutine线程数，默认20
		NotLoadCacheAtStart:  true,                        // 在启动时不读取本地缓存数据，true--不读取，false--读取
		UpdateCacheWhenEmpty: true,                        // 当服务列表为空时是否更新本地缓存，true--更新,false--不更新
		LogLevel:             nacosConfig.LogLevel,
		Username:             nacosConfig.Username,
		Password:             password,
	}

	serverConfigs := make([]constant.ServerConfig, 0)
	addrs := strings.Split(nacosConfig.Addr, ",")

	for _, addr := range addrs {
		if !strings.Contains(addr, "://") {
			addr = "http://" + addr
		}

		u, err := url.Parse(addr)
		if err != nil {
			return nil, nil, fmt.Errorf("parse addr error: %s, addr: %s\n", err.Error(), addr)
		}
		cfg := constant.ServerConfig{
			Scheme:      u.Scheme,
			IpAddr:      u.Hostname(),
			Port:        nacosConfig.Port,
			ContextPath: nacosConfig.Path,
		}
		if len(cfg.Scheme) == 0 {
			cfg.Scheme = constant.DEFAULT_SERVER_SCHEME
		}
		if len(cfg.ContextPath) == 0 {
			cfg.ContextPath = constant.DEFAULT_CONTEXT_PATH
		}
		if len(cfg.IpAddr) == 0 {
			cfg.IpAddr = u.Path
			if len(cfg.IpAddr) == 0 {
				return nil, nil, fmt.Errorf("parse addr error: ip is empty, addr: %s, host: %s, path: %s\n", addr, u.Host, u.Path)
			}
		}

		if ips, err := net.LookupIP(cfg.IpAddr); nil != err {
			return nil, nil, err
		} else if len(ips) == 0 {
			return nil, nil, fmt.Errorf("lookup ip fail, no ip found")
		} else {
			cfg.IpAddr = ips[0].String()
		}

		if len(u.Port()) > 0 {
			port, err := strconv.Atoi(u.Port())
			if err != nil {
				return nil, nil, fmt.Errorf("parse port error: %s, addr: %s, port: %s\n", err.Error(), addr, u.Port())
			}
			cfg.Port = uint64(port)
		}
		serverConfigs = append(serverConfigs, cfg)
	}

	if 0 == len(serverConfigs) {
		return nil, nil, fmt.Errorf("no valid nacos server address, addrs: %v", addrs)
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, nil, err
	}

	nammingClient, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})

	if err != nil {
		return nil, nil, err
	}

	return nammingClient, configClient, nil
}

func NewClusterService(nacosConfig *Config) (*ClusterService, error) {
	s := &ClusterService{
		localIp:   utils.GetLocalIp(),
		localPort: int(nacosConfig.ListenPort),
		instances: sync.Map{},
	}

	s.nacosConfig = nacosConfig

	client, config, err := createNacosClientHighAvailability(nacosConfig, nacosConfig.Password)
	if nil != err {
		return nil, errors.Wrap(err, "create naming client error")
	}

	s.Weight = defaultWeight
	s.localAddress = fmt.Sprintf("%s:%d", s.localIp, s.nacosConfig.ListenPort)

	// 注册自己
	s.registerParam = vo.RegisterInstanceParam{
		Ip:          s.localIp,
		Port:        s.nacosConfig.ListenPort,
		Weight:      s.Weight,
		Enable:      true,
		Healthy:     true,
		ClusterName: s.nacosConfig.Cluster,
		ServiceName: s.nacosConfig.Service,
		GroupName:   s.nacosConfig.Group,
		Ephemeral:   true,
		Metadata:    map[string]string{},
	}

	s.deregisterParam = vo.DeregisterInstanceParam{
		Ip:          s.localIp,
		Port:        s.nacosConfig.ListenPort,
		Cluster:     s.nacosConfig.Cluster,
		ServiceName: s.nacosConfig.Service,
		GroupName:   s.nacosConfig.Group,
		Ephemeral:   true,
	}

	// register
	if success, err := client.RegisterInstance(s.registerParam); nil != err {
		return nil, errors.Wrap(err, "register instance error")
	} else if !success {
		return nil, fmt.Errorf("register instance fail")
	}

	if nil == s.subscribedServices {
		s.subscribedServices = make(map[string]*EventHandler)
	}

	s.selfInstanceId = s.GetHostname()

	s.client = client
	s.config = config

	return s, nil
}

func (s *ClusterService) updateInstanceList(serviceName string, instances []InstanceInfo) {
	s.instances.Store(serviceName, instances)
}

// 获取一个服务的实例列表
func (s *ClusterService) GetInstances(serviceName string) []InstanceInfo {
	if val, ok := s.instances.Load(serviceName); !ok || nil == val {
		return nil
	} else {
		return val.([]InstanceInfo)
	}
}

// SelectOneInstance 使用加权循环调度算法（WRR）选择出一个服务的实例
func (s *ClusterService) SelectOneInstance(serviceName string) (*InstanceInfo, error) {
	return s.SelectOneInstanceFromGroup(serviceName, s.nacosConfig.Group)
}

func (s *ClusterService) SelectOneInstanceFromGroup(serviceName, groupName string) (*InstanceInfo, error) {
	if !s.IsReady() {
		return nil, errors.New("cluster service is not ready")
	}

	if ins, err := s.client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		Clusters:    []string{s.nacosConfig.Cluster},
		ServiceName: serviceName,
		GroupName:   groupName,
	}); nil != err {
		return nil, err
	} else {
		return &InstanceInfo{
			InstanceId:  ins.InstanceId,
			ServiceName: ins.ServiceName,
			Group:       groupName,
			ClusterName: ins.ClusterName,
			Ip:          ins.Ip,
			Port:        ins.Port,
			Weight:      ins.Weight,
			Enable:      ins.Enable,
		}, nil
	}
}

func decodeConfigValue(node *yaml.Node) map[string]string {
	var configs = map[string]string{}

	if yaml.DocumentNode == node.Kind && 0 != len(node.Content) {
		node = node.Content[0]
	}

	if yaml.MappingNode == node.Kind && len(node.Content) > 0 {
		var key string
		for i, d := range node.Content {
			if 0 == i%2 {
				key = d.Value
			} else {
				configs[key] = d.Value
			}
		}
	}
	return configs
}

// 获取一个配置数据
func (s *ClusterService) GetConfigMap(dataId string) (map[string]string, error) {
	if !s.IsReady() {
		return nil, errors.New("must call Init first")
	}

	str, err := s.config.GetConfig(
		vo.ConfigParam{
			DataId: dataId,
			Group:  s.nacosConfig.Group,
		},
	)

	if nil != err {
		return nil, err
	}

	var data yaml.Node
	err = yaml.Unmarshal([]byte(str), &data)
	if nil != err {
		return nil, err
	}

	return decodeConfigValue(&data), nil
}

func (s *ClusterService) GetConfig(dataId string, cfg interface{}) error {
	if !s.IsReady() {
		return errors.New("must call Init first")
	}

	str, err := s.config.GetConfig(
		vo.ConfigParam{
			DataId: dataId,
			Group:  s.nacosConfig.Group,
		},
	)

	if nil != err {
		return err
	}

	return yaml.Unmarshal([]byte(str), cfg)
}

// 注册一个监听
func (s *ClusterService) RegisterMonitor(serviceName string, monitor ClusterMonitor) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if nil == monitor {
		return errors.New("param error")
	}

	if !s.IsReady() {
		return errors.New("must call Init first")
	}

	if hdl, ok := s.subscribedServices[serviceName]; !ok {
		hdl = createEventHandler(serviceName, s.nacosConfig.Group, s)
		hdl.addMonitor(monitor)
		// 自己的服务已经订阅过了
		if err := s.client.Subscribe(&vo.SubscribeParam{
			ServiceName:       hdl.serviceName,
			GroupName:         hdl.groupName,
			Clusters:          []string{s.nacosConfig.Cluster},
			SubscribeCallback: hdl.onEvent,
		}); nil != err {
			return err
		}
		s.subscribedServices[serviceName] = hdl
	} else {
		hdl.addMonitor(monitor)
	}

	return nil
}

// Close 关闭服务
func (s *ClusterService) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	defer func() {
		s.client = nil
		s.config = nil
		s.subscribedServices = nil
	}()

	if s.IsReady() {
		if nil != s.subscribedServices {
			for _, hdl := range s.subscribedServices {
				hdl.setBlock(true)
				_ = s.client.Unsubscribe(&vo.SubscribeParam{
					ServiceName:       hdl.serviceName,
					Clusters:          []string{s.nacosConfig.Cluster},
					GroupName:         hdl.groupName,
					SubscribeCallback: hdl.onEvent,
				})
			}
		}

		if _, err := s.client.DeregisterInstance(s.deregisterParam); err != nil {
			return err
		}
	}

	return nil
}
