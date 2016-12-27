package smartyaddress

import (
	"google.golang.org/appengine/urlfetch"
	"net/url"
	"settings"
	"io/ioutil"
	"google.golang.org/appengine/log"
	"net/http"
	"errors"
	"golang.org/x/net/context"
	"github.com/dustin/gojson"
)

var ADDRESS_NOT_FOUND_ERROR error = errors.New("Address not found")

type strategy string

type Lookup struct {
	Street        string   `json:"street,omitempty"`
	Street2       string   `json:"street2,omitempty"`
	Secondary     string   `json:"secondary,omitempty"`
	City          string   `json:"city,omitempty"`
	State         string   `json:"state,omitempty"`
	ZIPCode       string   `json:"zipcode,omitempty"`
	LastLine      string   `json:"lastline,omitempty"`
	Addressee     string   `json:"addressee,omitempty"`
	Urbanization  string   `json:"urbanization,omitempty"`
	InputID       string   `json:"input_id,omitempty"`
	MaxCandidates int      `json:"candidates,omitempty"` // Default value: 1
	MatchStrategy strategy `json:"match,omitempty"`

	Results []*Candidate `json:"results,omitempty"`
}
// Candidate contains all output fields defined here:
// https://smartystreets.com/docs/us-street-api#http-response-output
type Candidate struct {
	InputID              string     `json:"input_id,omitempty"`
	InputIndex           int        `json:"input_index"`
	CandidateIndex       int        `json:"candidate_index"`
	Addressee            string     `json:"addressee,omitempty"`
	DeliveryLine1        string     `json:"delivery_line_1,omitempty"`
	DeliveryLine2        string     `json:"delivery_line_2,omitempty"`
	LastLine             string     `json:"last_line,omitempty"`
	DeliveryPointBarcode string     `json:"delivery_point_barcode,omitempty"`
	Components           Components `json:"components,omitempty"`
	Metadata             Metadata   `json:"metadata,omitempty"`
	Analysis             Analysis   `json:"analysis,omitempty"`
}

// Components contains all output fields defined here:
// https://smartystreets.com/docs/us-street-api#components
type Components struct {
	PrimaryNumber            string `json:"primary_number,omitempty"`
	StreetPredirection       string `json:"street_predirection,omitempty"`
	StreetName               string `json:"street_name,omitempty"`
	StreetPostdirection      string `json:"street_postdirection,omitempty"`
	StreetSuffix             string `json:"street_suffix,omitempty"`
	SecondaryNumber          string `json:"secondary_number,omitempty"`
	SecondaryDesignator      string `json:"secondary_designator,omitempty"`
	ExtraSecondaryNumber     string `json:"extra_secondary_number,omitempty"`
	ExtraSecondaryDesignator string `json:"extra_secondary_designator,omitempty"`
	PMBNumber                string `json:"pmb_number,omitempty"`
	PMBDesignator            string `json:"pmb_designator,omitempty"`
	CityName                 string `json:"city_name,omitempty"`
	DefaultCityName          string `json:"default_city_name,omitempty"`
	StateAbbreviation        string `json:"state_abbreviation,omitempty"`
	ZIPCode                  string `json:"zipcode,omitempty"`
	Plus4Code                string `json:"plus4_code,omitempty"`
	DeliveryPoint            string `json:"delivery_point,omitempty"`
	DeliveryPointCheckDigit  string `json:"delivery_point_check_digit,omitempty"`
	Urbanization             string `json:"urbanization,omitempty"`
}

// Metadata contains all output fields defined here:
// https://smartystreets.com/docs/us-street-api#metadata
type Metadata struct {
	RecordType               string  `json:"record_type,omitempty"`
	ZIPType                  string  `json:"zip_type,omitempty"`
	CountyFIPS               string  `json:"county_fips,omitempty"`
	CountyName               string  `json:"county_name,omitempty"`
	CarrierRoute             string  `json:"carrier_route,omitempty"`
	CongressionalDistrict    string  `json:"congressional_district,omitempty"`
	BuildingDefaultIndicator string  `json:"building_default_indicator,omitempty"`
	RDI                      string  `json:"rdi,omitempty"`
	ELOTSequence             string  `json:"elot_sequence,omitempty"`
	ELOTSort                 string  `json:"elot_sort,omitempty"`
	Latitude                 float64 `json:"latitude,omitempty"`
	Longitude                float64 `json:"longitude,omitempty"`
	Precision                string  `json:"precision,omitempty"`
	TimeZone                 string  `json:"time_zone,omitempty"`
	UTCOffset                float32 `json:"utc_offset,omitempty"`
	DST                      bool    `json:"dst,omitempty"`
}

// Analysis contains all output fields defined here:
// https://smartystreets.com/docs/us-street-api#analysis
type Analysis struct {
	DPVMatchCode      string `json:"dpv_match_code,omitempty"`
	DPVFootnotes      string `json:"dpv_footnotes,omitempty"`
	DPVCMRACode       string `json:"dpv_cmra,omitempty"`
	DPVVacantCode     string `json:"dpv_vacant,omitempty"`
	Active            string `json:"active,omitempty"`
	Footnotes         string `json:"footnotes,omitempty"` // https://smartystreets.com/docs/us-street-api#footnotes
	LACSLinkCode      string `json:"lacslink_code,omitempty"`
	LACSLinkIndicator string `json:"lacslink_indicator,omitempty"`
	SuiteLinkMatch    bool   `json:"suitelink_match,omitempty"`
	EWSMatch          bool   `json:"ews_match,omitempty"`
}

func CheckUSAddress(ctx context.Context, lookup *Lookup) (*Candidate, error) {
	u, err := url.Parse("https://us-street.api.smartystreets.com/street-address")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("auth-id", settings.SMARTYSTREETS_AUTH_ID)
	q.Add("auth-token", settings.SMARTYSTREETS_AUTH_TOKEN)
	q.Add("street", lookup.Street)
	q.Add("street2", lookup.Street2)
	q.Add("city", lookup.City)
	q.Add("state", lookup.State)
	q.Add("zipcode", lookup.ZIPCode)

	u.RawQuery = q.Encode()
	uri := u.String()
	log.Infof(ctx, "Requesting uri: %s", uri)
	r, err := urlfetch.Client(ctx).Get(uri)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
	bts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	log.Infof(ctx, "Check Address Response: %s", bts)
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}

	results := make([]*Candidate, 1)
	err = json.Unmarshal(bts, &results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ADDRESS_NOT_FOUND_ERROR
	}

	return results[0], nil
}
