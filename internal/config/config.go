package config

type SysConfig struct {
	Service  ServiceInfo `json:"service" yaml:"service"`
	Database MySqlConfig `json:"database" yaml:"database"`
	Consul   ConsulInfo  `json:"consul" yaml:"consul"`
}

type ServiceInfo struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Listen  string `json:"listen" yaml:"listen"`
	MaxQps  int    `json:"max_qps" yaml:"max_qps"`
}

type ConsulInfo struct {
	Addr             string   `json:"addr" yaml:"addr"`
	Port             uint     `json:"port" yaml:"port"`
	Prefix           string   `json:"prefix" yaml:"prefix"`
	Timeout          int32    `json:"timeout" yaml:"timeout"`
	RegisterInterval uint     `json:"register_interval" yaml:"register_interval"`
	RegisterTtl      uint     `json:"register_ttl" yaml:"register_ttl"`
	Token            string   `json:"token" yaml:"token"`
	RegistryAddrs    []string `json:"registry_addrs" yaml:"registry_addrs"`
}

type MySqlConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int64  `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	Loc      string `json:"loc" yaml:"loc"`
	Charset  string `json:"charset" yaml:"charset"`
}
