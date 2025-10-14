package prom

type Config struct {
	Namespace string `mapstructure:"namespace" yaml:"namespace"`
	//ApiUrls   string `mapstructure:"api_urls" yaml:"api_urls"`
	//PushGatewayUrl            string `mapstructure:"push_gateway_url" yaml:"push_gateway_url` // deprecated 使用 PushGatewayUrls
	PushGatewayUrls               string `mapstructure:"push_gateway_urls" yaml:"push_gateway_urls"`
	PushGatewayJob                string `mapstructure:"push_gateway_job" yaml:"push_gateway_job"`
	PushGatewayDurationSecond     int    `mapstructure:"push_gateway_duration_second" yaml:"push_gateway_duration_second"`
	PushGatewayDefaultActiveIndex *int   `mapstructure:"push_gateway_default_active_index" yaml:"push_gateway_default_active_index"`
}
