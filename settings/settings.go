package settings

import (
	"os"
	"net/http"
	"fmt"
	"strconv"
	"google.golang.org/appengine/log"
	"github.com/jcarm010/kodimerce/entities"
	"golang.org/x/net/context"
)

var (
	globalSettings *entities.ServerSettings
)

func GetAndReloadGlobalSettings(ctx context.Context) entities.ServerSettings {
	dbSettings, err := entities.GetServerSettings(ctx)
	if err != nil {
		log.Errorf(ctx, "Error getting stored server settings: %s", err)
		envSettings := getEnvSettings()
		dbSettings = &envSettings
		if err == entities.ErrSettingsNotFound {
			err = entities.StoreServerSettings(ctx, dbSettings)
			if err != nil {
				log.Errorf(ctx, "Error storing server settings: %s", err)
			}
		}
	}

	globalSettings = dbSettings
	return *dbSettings
}

func GetGlobalSettings(ctx context.Context) entities.ServerSettings {
	if globalSettings != nil {
		return *globalSettings
	}

	return GetAndReloadGlobalSettings(ctx)
}

func getEnvSettings() entities.ServerSettings {
	taxPercent, _ := strconv.ParseFloat(os.Getenv("TAX_PERCENT"), 64)
	wwwRedirect, _ := strconv.ParseBool(os.Getenv("WWW_REDIRECT"))

	return  entities.ServerSettings {
		Author: os.Getenv("AUTHOR"),
		CompanyName: os.Getenv("COMPANY_NAME"),
		CompanyNameAlternate: os.Getenv("COMPANY_NAME_ALTERNATE"),
		CompanyMailingAddress: os.Getenv("COMPANY_MAILING_ADDRESS"),
		CompanyContactEmail: os.Getenv("COMPANY_CONTACT_EMAIL"),
		CompanyContactPhone: os.Getenv("COMPANY_CONTACT_PHONE"),
		CompanySupportEmail: os.Getenv("COMPANY_SUPPORT_EMAIL"),
		CompanyOrdersEmail: os.Getenv("COMPANY_ORDERS_EMAIL"),
		TaxPercent: taxPercent,
		CompanyUrl: os.Getenv("COMPANY_URL"),
		CompanyGoogleMapsUrl: os.Getenv("COMPANY_GOOGLE_MAPS_URL"),
		CompanyGoogleMapsEmbedUrl: os.Getenv("COMPANY_GOOGLE_MAPS_EMBED_URL"),
		FacebookUrl: os.Getenv("FACEBOOK_URL"),
		FacebookAppId: os.Getenv("FACEBOOK_APP_ID"),
		InstagramUrl: os.Getenv("INSTAGRAM_URL"),
		TwitterUrl: os.Getenv("TWITTER_URL"),
		LinkedInUrl: os.Getenv("LINKEDIN_URL"),
		YoutubeUrl: os.Getenv("YOUTUBE_URL"),
		TwitterHandle: os.Getenv("TWITTER_HANDLE"),
		GoogleAnalyticsAccountId: os.Getenv("GOOGLE_ANALYTICS_ACCOUNT_ID"),
		GoogleTagManagerId: os.Getenv("GOOGLE_TAG_MANAGER_ID"),

		PayPalEnvironment: os.Getenv("PAYPAL_ENVIRONMENT"),
		PayPalApiUrl: os.Getenv("PAYPAL_API_URL"),
		PayPalEmail: os.Getenv("PAYPAL_EMAIL"),
		PayPalAccount: os.Getenv("PAYPAL_ACCOUNT"),
		PayPalApiClientId: os.Getenv("PAYPAL_API_CLIENT_ID"),
		PayPalApiClientSecret: os.Getenv("PAYPAL_API_CLIENT_SECRET"),
		PayPalAllowedPaymentOption: os.Getenv("PAYPAL_ALLOWED_PAYMENT_OPTION"), //posible: UNRESTRICTED, INSTANT_FUNDING_SOURCE, IMMEDIATE_PAY
		PayPalNoteToPayer: os.Getenv("PAYPAL_NOTE_TO_PAYER"),

		SmartyStreetsAuthId: os.Getenv("SMARTYSTREETS_AUTH_ID"),
		SmartyStreetsAuthToken: os.Getenv("SMARTYSTREETS_AUTH_TOKEN"),

		EmailSender: os.Getenv("EMAIL_SENDER"),
		SendGridKey: os.Getenv("SENDGRID_KEY"),

		MetaTitleHome: os.Getenv("META_TITLE_HOME"),

		MetaDescriptionHome: os.Getenv("META_DESCRIPTION_HOME"),
		MetaDescriptionStore: os.Getenv("META_DESCRIPTION_STORE"),
		MetaDescriptionReferrals: os.Getenv("META_DESCRIPTION_REFERRALS"),
		MetaDescriptionContact: os.Getenv("META_DESCRIPTION_CONTACT"),
		MetaDescriptionCart: os.Getenv("META_DESCRIPTION_CART"),
		MetaDescriptionBlog: os.Getenv("META_DESCRIPTION_BLOG"),
		MetaDescriptionGalleries: os.Getenv("META_DESCRIPTION_GALLERIES"),

		DescriptionBlogABout: os.Getenv("DESCRIPTION_BLOG_ABOUT"),
		WwwRedirect: wwwRedirect,
	}
}


func ServerUrl(r *http.Request) string {
	httpHeader := "http"
	if r.TLS != nil {
		httpHeader = "https"
	}

	return fmt.Sprintf("%s://%s", httpHeader, r.Host)
}