package config

type SharedConfig struct {
	JwtSecret []byte
}

var sharedConfig *SharedConfig

func GetSharedConfig() *SharedConfig {

	if sharedConfig != nil {
		return sharedConfig
	}

	sharedConfig := &SharedConfig{
		JwtSecret: []byte("mozogooloo"),
	}

	return sharedConfig
}
