package webserver

// Config ...
type Config struct {
	Addr string    `json:"address"`
	Port int       `json:"port"`
	TLS  TLSConfig `json:"tls"`
}

// TLSConfig ...
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"certPath"`
	KeyFile  string `json:"keyPath"`
}
