package entities

type ServerConfig struct {
	ConfigKey string
	CompanyName string `datastore:",noindex"`
	CompanyAddress string `datastore:",noindex"`
	CompanyEmail string `datastore:",noindex"`
	CompanyPhone string `datastore:",noindex"`
}

func NewServerConfig(companyName string, companyAddress string, companyEmail string, companyPhone string) ServerConfig {
	return ServerConfig{
		ConfigKey: SERVER_CONFIG_KEY,
		CompanyName: companyName,
		CompanyAddress: companyAddress,
		CompanyEmail: companyEmail,
		CompanyPhone: companyPhone,
	}
}
