package config

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
	RoleGuest = "guest"
)

var RoleDriveSize = map[string]int64{
	RoleAdmin: 30 * 1024 * 1024 * 1024, // 30 GB for admin
	RoleUser:  5 * 1024 * 1024 * 1024,  // 5 GB for regular users
	RoleGuest: 1 * 1024 * 1024 * 1024,  // 1 GB for guests
}

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
