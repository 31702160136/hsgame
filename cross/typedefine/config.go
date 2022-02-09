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
	ServerId     int    `json:"server_id"`
}

type HttpResult struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
