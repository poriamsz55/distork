package config

type ConfigDB struct {
	DatabaseName string
	DistorkColl  string
	UserColl     string
	RoomColl     string
}

var (
	configDB *ConfigDB
)

// GetConfig returns the instance of ConfigDB, loading it if it has not been loaded before
func GetConfigDB() *ConfigDB {

	if configDB != nil {
		return configDB
	}

	configDB = &ConfigDB{
		DatabaseName: "distork",
		DistorkColl:  "distork",
		UserColl:     "users",
		RoomColl:     "rooms",
	}
	return configDB
}
