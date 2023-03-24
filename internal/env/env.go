package env

type Env struct {
	IntervalSeconds  string `envconfig:"interval_seconds" required:"true"`
	DatabaseFilepath string `envconfig:"database_filepath" required:"true"`
	Port             string `envconfig:"port" required:"true"`
	Debug            bool   `envconfig:"debug" required:"true"`
}
