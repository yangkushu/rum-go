package elasticsearch

type Config struct {
	Addresses          string `mapstructure:"addresses"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	IndexPrefix        string `mapstructure:"index_prefix"`
	EnableLogger       bool   `mapstructure:"enable_logger"`
	EnableRequestBody  bool   `mapstructure:"enable_request_body"`
	EnableResponseBody bool   `mapstructure:"enable_response_body"`
	Scheme             string `mapstructure:"scheme"`
}
