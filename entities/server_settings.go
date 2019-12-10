package entities

import (
	"github.com/jcarm010/kodimerce/datastore"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var (
	ErrSettingsNotFound = errors.New("not found")
)

type ServerSettings struct {
	Author                    string  `json:"author"`
	CompanyName               string  `json:"company_name"`
	CompanyNameAlternate      string  `json:"company_name_alternate"`
	CompanyMailingAddress     string  `json:"company_mailing_address"`
	CompanyContactEmail       string  `json:"company_contact_email"`
	CompanyContactPhone       string  `json:"company_contact_phone"`
	CompanySupportEmail       string  `json:"company_support_email"`
	CompanyOrdersEmail        string  `json:"company_orders_email"`
	TaxPercent                float64 `json:"tax_percent"`
	CompanyUrl                string  `json:"company_url"`
	CompanyGoogleMapsUrl      string  `json:"company_google_maps_url"`
	CompanyGoogleMapsEmbedUrl string  `json:"company_google_maps_embed_url"`
	FacebookUrl               string  `json:"facebook_url"`
	FacebookAppId             string  `json:"facebook_app_id"`
	InstagramUrl              string  `json:"instagram_url"`
	TwitterUrl                string  `json:"twitter_url"`
	LinkedInUrl               string  `json:"linked_in_url"`
	YoutubeUrl                string  `json:"youtube_url"`
	TwitterHandle             string  `json:"twitter_handle"`
	GoogleAnalyticsAccountId  string  `json:"google_analytics_account_id"`
	GoogleTagManagerId        string  `json:"google_tag_manager_id"`

	FareHarborShortName string `json:"fareharbor_short_name"`

	PayPalEnvironment          string `json:"pay_pal_environment"`
	PayPalApiUrl               string `json:"pay_pal_api_url"`
	PayPalEmail                string `json:"pay_pal_email"`
	PayPalAccount              string `json:"pay_pal_account"`
	PayPalApiClientId          string `json:"pay_pal_api_client_id"`
	PayPalApiClientSecret      string `json:"pay_pal_api_client_secret"`
	PayPalAllowedPaymentOption string `json:"pay_pal_allowed_payment_option"` //posible: UNRESTRICTED, INSTANT_FUNDING_SOURCE, IMMEDIATE_PAY
	PayPalNoteToPayer          string `json:"pay_pal_note_to_payer"`

	SmartyStreetsAuthId    string `json:"smarty_streets_auth_id"`
	SmartyStreetsAuthToken string `json:"smarty_streets_auth_token"`

	SMTPServer   string `json:"smtp_server"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUserName string `json:"smtp_user_name"`
	SMTPPassword string `json:"smtp_password"`

	EmailSender string `json:"email_sender"`
	SendGridKey string `json:"send_grid_key"`

	MetaTitleHome string `json:"meta_title_home"`

	MetaDescriptionHome      string `json:"meta_description_home"`
	MetaDescriptionStore     string `json:"meta_description_store"`
	MetaDescriptionReferrals string `json:"meta_description_referrals"`
	MetaDescriptionContact   string `json:"meta_description_contact"`
	MetaDescriptionCart      string `json:"meta_description_cart"`
	MetaDescriptionBlog      string `json:"meta_description_blog"`
	MetaDescriptionGalleries string `json:"meta_description_galleries"`

	DescriptionBlogABout string `json:"description_blog_about"`
	WwwRedirect          bool   `json:"www_redirect"`
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
