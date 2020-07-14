package app

type AppConfig struct {
	Log      Log    `yaml:"log"`
	API      API    `yaml:"api"`
	DB       DB     `yaml:"db"`
}
type Log struct {
	File  string `yaml:"file"`
}

type API struct {
	GRPCPort string `yaml:"grpcport"`
}

type DB struct {
	DSN    string `yaml:"dsn"`
	Dialect string  `yaml:"dialect"`
}
