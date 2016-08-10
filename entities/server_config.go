package entities

type ServerConfig struct {
	ConfigKey string
	PrimaryColor string `datastore:",noindex"`
	SuccessColor string `datastore:",noindex"`
	WarningColor string `datastore:",noindex"`
	DangerColor string `datastore:",noindex"`
	DefaultColor string `datastore:",noindex"`
	PrimaryFontColor string `datastore:",noindex"`
	SuccessFontColor string `datastore:",noindex"`
	WarningFontColor string `datastore:",noindex"`
	DangerFontColor string `datastore:",noindex"`
	DefaultFontColor string `datastore:",noindex"`
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		PrimaryColor: "#3f51b5",
		SuccessColor: "#3fb5a3",
		WarningColor: "#b5a33f",
		DangerColor: "#b53f51",
		DefaultColor: "#ffffff",
		PrimaryFontColor: "#ffffff",
		SuccessFontColor: "#ffffff",
		WarningFontColor: "#ffffff",
		DangerFontColor: "#ffffff",
		DefaultFontColor: "#000000",
	}
}