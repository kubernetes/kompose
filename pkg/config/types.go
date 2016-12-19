package config

type Profile struct {
	Provider string   `yaml:"provider"`
	Objects  []string `yaml:"objects"`
}

type ProfilesMap map[string]Profile

type Config struct {
	Profiles       ProfilesMap `yaml:"profiles,omitempty"`
	CurrentProfile string      `yaml:"current-profile,omitempty"`
}
