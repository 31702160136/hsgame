package typedefine

type Config struct {
	Port         int
	HttpSpanPort int
	DBPort       int
	DBHost       string
	DBName       string
	DBAuthSource string
	DBUserName   string
	DBPassword   string
	Log          string
	SignatureKey string
}
