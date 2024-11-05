package config

type ConfigDrive struct {
	UploadDir string
}

var (
	configDrive *ConfigDrive
)

// GetConfigDrive returns the instance of Config, loading it if it has not been loaded before
func GetConfigDrive() *ConfigDrive {

	if configDrive != nil {
		return configDrive
	}

	configDrive = &ConfigDrive{
		UploadDir: "uploads",
	}
	return configDrive
}
