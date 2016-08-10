package entities

type CompanyConfig struct {
	ConfigKey string
	CompanyName string `datastore:",noindex"`
	CompanyAddress string `datastore:",noindex"`
	CompanyEmail string `datastore:",noindex"`
	CompanyPhone string `datastore:",noindex"`
}

func NewCompanyConfig(companyName string, companyAddress string, companyEmail string, companyPhone string) CompanyConfig {
	return CompanyConfig{
		ConfigKey: CONFIG_KEY_COMPANY,
		CompanyName: companyName,
		CompanyAddress: companyAddress,
		CompanyEmail: companyEmail,
		CompanyPhone: companyPhone,
	}
}
