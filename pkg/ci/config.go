package ci

type Config struct {
	Runner   string   `yaml:"runner_image"`
	Commands []string `yaml:"commands"`
}

const ConfigPath = "master:.kek/config.yml"
