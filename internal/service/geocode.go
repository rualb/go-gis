package service

import (
	"encoding/json"
	"fmt"
	"go-gis/internal/config"
	"go-gis/internal/tool/toolhttp"
	xlog "go-gis/internal/tool/toollog"
	"net/url"
	"strings"
)

// respItemGeocodeOSM use as array
type respItemGeocodeOSM struct {
	DisplayName string `json:"display_name"`
}

type respItemGeocodeGMAPS struct {
	FormattedAddress string `json:"formatted_address"`
}
type respGeocodeGMAPS struct {
	Results []respItemGeocodeGMAPS `json:"results"`
}

type GeocodeService interface {
	LocationToAddress(latLng string, lang string) (address string, err error)
}

type defaultGeocodeSrv struct {
	appConfig *config.AppConfig
	Debug     bool
}

func (x *defaultGeocodeSrv) LocationToAddress(latLng string, lang string) (address string, err error) {

	if lang == "" {
		lang = "en"
	}

	address, err = x.locationToAddressOSM(latLng, lang)
	if err != nil {
		return "", err
	}
	if address != "" {
		return address, nil
	}

	address, err = x.locationToAddressGMAPS(latLng, lang)
	if err != nil {
		return "", err
	}
	if address != "" {
		return address, nil
	}

	xlog.Error("Define geocode service")

	return "", fmt.Errorf("error no any geocode service")
}

func (x *defaultGeocodeSrv) locationToAddressOSM(latLng string, lang string) (address string, err error) {
	cnf := &x.appConfig.OsmGateway
	if !cnf.Enabled {
		return "", nil
	}
	latLng = url.QueryEscape(latLng)
	lang = url.QueryEscape(lang)
	apiKey := url.QueryEscape(cnf.APIKey)

	baseURL := cnf.URL

	baseURL = strings.ReplaceAll(baseURL, "{LatLng}", latLng)
	baseURL = strings.ReplaceAll(baseURL, "{Lang}", lang)
	baseURL = strings.ReplaceAll(baseURL, "{ApiKey}", apiKey)

	data, err := toolhttp.GetBytes(baseURL, nil, map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; AcmeInc/1.0)",
	})

	if err != nil {
		xlog.Error("OSM connect: %v", err)
		return "", fmt.Errorf("error on geocode")
	}

	respObj := []respItemGeocodeOSM{} // array
	err = json.Unmarshal(data, &respObj)
	if err != nil {
		xlog.Error("OSM resp: %v", err)
		return "", fmt.Errorf("error on geocode")
	}

	respItems := respObj

	if len(respItems) == 0 {
		address = "" // undef
	} else {
		address = respObj[0].DisplayName
	}

	return address, err
}
func (x *defaultGeocodeSrv) locationToAddressGMAPS(latLng string, lang string) (address string, err error) {
	cnf := &x.appConfig.GmapsGateway
	if !cnf.Enabled {
		return "", nil
	}
	latLng = url.QueryEscape(latLng)
	lang = url.QueryEscape(lang)
	apiKey := url.QueryEscape(cnf.APIKey)

	baseURL := cnf.URL

	baseURL = strings.ReplaceAll(baseURL, "{LatLng}", latLng)
	baseURL = strings.ReplaceAll(baseURL, "{Lang}", lang)
	baseURL = strings.ReplaceAll(baseURL, "{ApiKey}", apiKey)

	data, err := toolhttp.GetBytes(baseURL, nil, map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; AcmeInc/1.0)",
	})

	if err != nil {
		xlog.Error("GMAPS connect: %v", err)
		return "", fmt.Errorf("error on geocode")
	}

	respObj := respGeocodeGMAPS{} // array
	err = json.Unmarshal(data, &respObj)
	if err != nil {
		xlog.Error("GMAPS resp: %v", err)
		return "", fmt.Errorf("error on geocode")
	}

	respItems := respObj.Results

	if len(respItems) == 0 {
		address = "" // undef
	} else {
		address = respItems[0].FormattedAddress
	}

	return address, err
}

func NewGeocode(appConfig *config.AppConfig) GeocodeService {

	return &defaultGeocodeSrv{
		Debug:     appConfig.Debug,
		appConfig: appConfig,
	}

}
