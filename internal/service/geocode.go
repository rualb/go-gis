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
		return "", fmt.Errorf("error on OSM connect: %v", err)
	}

	respObj := []respItemGeocodeOSM{} // array
	err = json.Unmarshal(data, &respObj)
	if err != nil {
		return "", fmt.Errorf("error on OSM resp: %v", err)
	}

	respItems := respObj

	if len(respItems) == 0 {
		address = "" // undef
	} else {
		address = respObj[0].DisplayName
		if cnf.Stdout {
			xlog.Info("Geocode: [LatLng: %v] [Address: %v]", latLng, address)
		}
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
		return "", fmt.Errorf("error on GMAPS connect: %v", err)
	}

	respObj := respGeocodeGMAPS{} // array
	err = json.Unmarshal(data, &respObj)
	if err != nil {

		return "", fmt.Errorf("error on GMAPS resp: %v", err)
	}

	respItems := respObj.Results

	if len(respItems) == 0 {
		address = "" // undef
	} else {
		address = respItems[0].FormattedAddress
		if cnf.Stdout {
			xlog.Info("Geocode: [LatLng: %v] [Address: %v]", latLng, address)
		}
	}

	return address, err
}

func NewGeocode(appConfig *config.AppConfig) GeocodeService {

	return &defaultGeocodeSrv{
		Debug:     appConfig.Debug,
		appConfig: appConfig,
	}

}
