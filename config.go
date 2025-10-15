package rum

import (
	"github.com/yangkushu/rum-go/elasticsearch"
	"github.com/yangkushu/rum-go/log"
	"github.com/yangkushu/rum-go/messagequeue"
	//"github.com/yangkushu/rum-go/nacos"
	"github.com/yangkushu/rum-go/postgres"
	"github.com/yangkushu/rum-go/redis"
)

type Config struct {
	Redis         *redis.Config         `mapstructure:"redis"`
	Elasticsearch *elasticsearch.Config `mapstructure:"elasticsearch"`
	Postgres      *postgres.Config      `mapstructure:"postgres"`
	//Nacos         *nacos.Config             `mapstructure:"nacos"`
	Log   *log.Config               `mapstructure:"log"`
	Kafka *messagequeue.KafkaConfig `mapstructure:"kafka"`
}
