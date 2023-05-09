package config

type ServerConfig struct {
	ServerName            string   `mapstructure:"server-name" json:"server-name" yaml:"server-name"`
	Desc                  string   `mapstruture:"desc" json:"desc" yaml:"desc"`
	DispatchURL           string   `mapstruture:"dispatch-url" json:"dispatch-url" yaml:"dispatch-url"`
	AgentRoute            string   `mapstruture:"agent-route" json:"agent-route" yaml:"agent-route"`
	RegisterInfoCacheUrl  string   `mapstruture:"register-info-cache-url" json:"register-info-cache-url" yaml:"register-info-cache-url"`
	SharedServiceDBConfig DBConfig `mapstruture:"shared-service-db-config" json:"shared-service-db-config" yaml:"shared-service-db-config"`
	DBConfig              DBConfig `mapstruture:"db-config" json:"db-config" yaml:"db-config"`
}

type DBConfig struct {
	Username string `mapstruture:"username" json:"username" yaml:"username"`
	Password string `mapstruture:"password" json:"password" yaml:"password"`
	Host     string `mapstruture:"host" json:"host" yaml:"host"`
	Port     string `mapstruture:"port" json:"port" yaml:"port"`
	Dbname   string `mapstruture:"db-name" json:"db-name" yaml:"db-name"`
}
