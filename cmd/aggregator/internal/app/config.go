package app

type Config struct {
	Log   Log   `yaml:"log"`
	DB    DB    `yaml:"db"`
	Queue Queue `yaml:"queue"`
	Kafka Kafka `yaml:"kafka"`
}
type Log struct {
	File string `yaml:"file"`
}

type DB struct {
	DSN     string `yaml:"dsn"`
	Dialect string `yaml:"dialect"`
}

type Queue struct {
	Name string `yaml:"name"`
}
type Kafka struct {
	Topic         string `yaml:"topic"`
	Addr          string `yaml:"addr"`
	ConsumerGroup string `yaml:"consumergroup"`
	MinSize       int    `yaml:"minsize"`
	MaxSize       int    `yaml:"maxsize"`
}
