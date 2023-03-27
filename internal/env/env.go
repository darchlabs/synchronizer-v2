package env

type Env struct {
	IntervalSeconds string `envconfig:"interval_seconds" required:"true"`
	DatabaseDSN     string `envconfig:"database_dsn" required:"true"`
	Port            string `envconfig:"port" required:"true"`
	Debug           bool   `envconfig:"debug" default:"false"`
}
