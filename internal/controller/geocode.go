package controller

// Handler web req handler

import (
	"go-gis/internal/config/consts"
	"go-gis/internal/service"
	xlog "go-gis/internal/util/utillog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type locationDTO struct {
	LatLng string `query:"lat_lng"` // not lat_lng but LatLng
	Lang   string `query:"lang"`
}

func (x locationDTO) validate() bool {

	// len(30) "123.1234567890, 123.1234567890"
	if len(x.LatLng) > consts.LocationTextSize {
		return false
	}

	// len(2)
	if len(x.Lang) > consts.LangTextSize {
		return false
	}

	return true
}

type addressDTO struct {
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
	dto := &locationDTO{}
	err := c.Bind(dto)
	if err != nil {
		return err
	}

	if !dto.validate() {
		return c.NoContent(http.StatusBadRequest)
	}

	g := x.appService.Geocode()

	addr, err := g.LocationToAddress(dto.LatLng, dto.Lang)
	if err != nil {
		xlog.Error("Gocode service error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, addressDTO{Address: addr})

}
