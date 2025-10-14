package config

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"strings"
)

// Loader 配置加载器
type Loader struct {
	v              *viper.Viper      // Viper实例
	setReadConfigs []func() error    // 读取配置的函数
	configPath     string            // 配置文件路径
	configNames    []string          // 配置文件名称
	configType     string            // 配置文件类型
	envMapper      map[string]string // 环境变量映射
	envPrefix      string            // 自动替换到配置文件的环境变量前缀
}

// NewConfigLoader 创建一个新的配置加载器实例
// configPath 和 configNames 分别指定配置文件的路径和名称。
// 支持多个配置文件，后面的配置文件会覆盖前面的配置文件。
func NewConfigLoader() *Loader {
	loader := &Loader{
		v: viper.New(),
		//configPath:  configPath,
		//configNames: configNames,
		//configType: "yaml",
		envPrefix: "GO", // 默认前缀
	}
	//v := loader.v

	//fmt.Printf("init config ...\n")

	// 如果.env存在，就加载到环境变量
	//_ = godotenv.Load()
	//err := godotenv.Load()
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		fmt.Println(".env file does not exist. Skipping .env load.")
	//	} else {
	//		// If the error is not because the file doesn't exist, log an actual error
	//		fmt.Printf("Error loading .env file: %v\n", err)
	//	}
	//} else {
	//	fmt.Println(".env file loaded successfully.")
	//}

	//// 读取环境变量
	//v.AutomaticEnv()
	//v.SetEnvPrefix("GO")
	//// 替换环境变量中的 . 为 _
	//v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return loader
}

func (l *Loader) SetConfigFileYaml(configPath string, configNames []string) {
	//fmt.Printf("init config, config path:%s,configNames:%s\n", configPath, configNames)
	l.configPath = configPath
	l.configNames = configNames
	l.configType = "yaml"

	// 设置配置文件类型和路径
	l.v.SetConfigType(l.configType) // example: "yaml"
	l.v.AddConfigPath(configPath)
}

func (l *Loader) Init() error {
	for _, configName := range l.configNames {
		// 这段代码有问题，无法正常检查文件是否存在，先注释掉
		// 如果配置文件不存在，就跳过
		//fullPath := l.configPath + "/" + configName /* + "." + l.configType*/
		//if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		//	log.Error("Config file does not exist.", log.String("fullPath", fullPath))
		//	continue
		//}

		// 读取配置文件
		l.v.SetConfigName(configName)
		err := l.v.MergeInConfig() // Use MergeInConfig instead of ReadConfig to allow multiple files
		if err != nil {
			return fmt.Errorf("load config error:merge config error: %w", err)
		}
	}

	// 通过其他方式读取配置
	for _, fn := range l.setReadConfigs {
		if err := fn(); err != nil {
			return fmt.Errorf("load config error:read config error: %w", err)
		}
	}

	// 加载 .env
	_ = godotenv.Load()

	// 替换环境变量中的 . 为 _
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if l.envPrefix != "" {
		// 读取环境变量
		l.v.AutomaticEnv()
		l.v.SetEnvPrefix(l.envPrefix)
	}

	// 通过环境变量映射设置配置
	for k, v := range l.envMapper {
		if err := l.v.BindEnv(k, v); err != nil {
			return fmt.Errorf("load config error:bind env error: %w", err)
		}
	}

	return nil
}

func (l *Loader) LoadConfig(configStruct any) error {
	// 将配置加载到结构体中
	if err := l.v.Unmarshal(configStruct); err != nil {
		return fmt.Errorf("load config error:unmarshal config error: %w", err)
	}

	//fmt.Printf("all env keys :" + fmt.Sprintf("%v\n", viper.AllKeys()))
	//fmt.Printf("config data: %+v\n", configStruct)
	return nil
}

// Load 将配置加载到指定的结构体中
// configStruct 是一个指向Config结构体的指针，用于存储加载的配置
// 需要在SetEnvMapper、SetReadConfig之后调用。理论上可以多次调用，但是没有测试过。
func (l *Loader) Load(configStruct any) error {
	if err := l.Init(); nil != err {
		return err
	}

	return l.LoadConfig(configStruct)
}

// SetReadConfig 读取配置，给nacos配置中心使用，需要在Load之前调用
func (l *Loader) SetReadConfig(content string) {
	fn := func() error {
		// 解析内容并赋给viper
		if err := viper.ReadConfig(bytes.NewBufferString(content)); err != nil {
			return err
		}
		return nil
	}
	l.setReadConfigs = append(l.setReadConfigs, fn)
}

// SetEnvMapper 设置环境变量映射,map[配置键]环境变量键，需要在Load之前调用
func (l *Loader) SetEnvMapper(envMapper map[string]string) {
	// 如果没有设置，就设置
	if len(l.envMapper) == 0 {
		l.envMapper = envMapper
	} else {
		// 如果已经设置了，就合并
		for k, v := range envMapper {
			l.envMapper[k] = v
		}
	}
}

func (l *Loader) SetEnvPrefix(envPrefix string) {
	l.envPrefix = envPrefix
	l.v.SetEnvPrefix(envPrefix)
}
