package controller

// Handler web req handler

import (
	"go-gis/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type LocationDto struct {
	LatLng string `query:"LatLng"` // not lat_lng but LatLng
	Lang   string `query:"Lang"`
}

type AddressDto struct {
	Address string `json:"address"`
}

// GeocodeController controller
type GeocodeController struct {
	appService service.AppService
	webCtxt    echo.Context
	Debug      bool
}

// NewGeocodeController new controller
func NewGeocodeController(appService service.AppService, c echo.Context) *GeocodeController {

	appConfig := appService.Config()
	return &GeocodeController{
		Debug:      appConfig.Debug,
		appService: appService,
		webCtxt:    c,
	}
}

// Geocode latlng to address
func (x *GeocodeController) Geocode() error {

	c := x.webCtxt
	dto := &LocationDto{}
	err := c.Bind(dto)
	if err != nil {
		return err
	}

	g := x.appService.Geocode()

	addr, err := g.LocationToAddress(dto.LatLng, dto.Lang)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, AddressDto{Address: addr})

}
