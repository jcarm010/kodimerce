package entities

import (
	"google.golang.org/appengine/datastore"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var (
	ErrSettingsNotFound = errors.New("not found")
)

type ServerSettings struct {
	Author string
	CompanyName string
	CompanyNameAlternate string
	CompanyMailingAddress string
	CompanyContactEmail string
	CompanyContactPhone string
	CompanySupportEmail string
	CompanyOrdersEmail string
	TaxPercent float64
	CompanyUrl string
	CompanyGoogleMapsUrl string
	CompanyGoogleMapsEmbedUrl string
	FacebookUrl string
	FacebookAppId string
	InstagramUrl string
	TwitterUrl string
	LinkedInUrl string
	YoutubeUrl string
	TwitterHandle string
	GoogleAnalyticsAccountId string
	GoogleTagManagerId string

	PayPalEnvironment string
	PayPalApiUrl string
	PayPalEmail string
	PayPalAccount string
	PayPalApiClientId string
	PayPalApiClientSecret string
	PayPalAllowedPaymentOption string //posible: UNRESTRICTED, INSTANT_FUNDING_SOURCE, IMMEDIATE_PAY
	PayPalNoteToPayer string

	SmartyStreetsAuthId string
	SmartyStreetsAuthToken string

	EmailSender string
	SendGridKey string

	MetaTitleHome string

	MetaDescriptionHome string
	MetaDescriptionStore string
	MetaDescriptionReferrals string
	MetaDescriptionContact string
	MetaDescriptionCart string
	MetaDescriptionBlog string
	MetaDescriptionGalleries string

	DescriptionBlogABout string
	WwwRedirect bool
}

func GetServerSettings(ctx context.Context) (*ServerSettings, error) {
	dbSettings := &ServerSettings{}
	key := datastore.NewKey(ctx, "server-settings", "active-settings", 0, nil)
	err := datastore.Get(ctx, key, dbSettings)
	if err == datastore.ErrNoSuchEntity {
		return nil, ErrSettingsNotFound
	}

	return dbSettings, err
}

func StoreServerSettings(ctx context.Context, serverSettings *ServerSettings) (error) {
	key := datastore.NewKey(ctx, "server-settings", "active-settings", 0, nil)
	_, err := datastore.Put(ctx, key, serverSettings)
	return err
}