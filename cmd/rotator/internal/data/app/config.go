package app

type AppConfig struct {
	Log   Log   `yaml:"log"`
	API   API   `yaml:"api"`
	DB    DB    `yaml:"db"`
	Algo  Algo  `yaml:"db"`
	Queue Queue `yaml:"queue"`
	Kafka Kafka `yaml:"kafka"`
}
type Log struct {
	File string `yaml:"file"`
}

type API struct {
	GRPCPort   string `yaml:"grpcport"`
	GRPCGWPort string `yaml:"grpcgwport"`
}

type DB struct {
	DSN     string `yaml:"dsn"`
	Dialect string `yaml:"dialect"`
}

type Algo struct {
	Name string `yaml:"name"`
}

type Queue struct {
	Name string `yaml:"name"`
}
type Kafka struct {
	Topic string `yaml:"topic"`
	Port  string `yaml:"port"`
}
