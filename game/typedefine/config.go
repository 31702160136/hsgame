package typedefine

type Config struct {
	Database     string
	SignatureKey string
	OptAddress   string
	Log          string
}

type ServerConfig struct {
	Name         string `json:"name"`
	Port         int    `json:"port"`
	HttpSpanPort int    `json:"http_span_port"`
	Nats         string `json:"nats"`
	CrossServer  int    `json:"cross_server"`
	MaxOnline    int    `json:"max_online"`
}

type HttpResult struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
