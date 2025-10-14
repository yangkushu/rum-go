package objectstorage

type S3Config struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	Region          string `mapstructure:"region" yaml:"region"`
	ForcePathStyle  bool   `mapstructure:"force_path_style" yaml:"force_path_style"`
}
