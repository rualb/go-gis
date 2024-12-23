package e2e

import (
	xcmd "go-gis/internal/cmd"
	"go-gis/internal/util/utilhttp"
	"os"
	"strings"
	"testing"
	"time"
)

// TestHealthController_Check_Stats tests the ?cmd=stats case in the Check method
func TestCmd(t *testing.T) {
	// Setup Echo context

	//   Trafalgar

	os.Setenv("APP_ENV", "testing")
	os.Setenv("APP_OSM_ENABLED", "1")
	os.Setenv("APP_OSM_STDOUT", "1")

	cmd := xcmd.Command{}

	go cmd.Exec()

	time.Sleep(3 * time.Second)

	urls := []struct {
		title  string
		url    string
		query  map[string]string
		search []string
	}{
		// http://127.0.0.1:31180/gis/api/geocode?lat_lng=51.50814,-0.12848&lang=en
		{title: "test loc to address", search: []string{`"address"`, "Trafalgar Square"}, url: "http://127.0.0.1:31180/gis/api/geocode", query: map[string]string{"lang": "en", "lat_lng": "51.50814,-0.12848"}},
	}

	for _, itm := range urls {

		t.Run(itm.title, func(t *testing.T) {

			t.Logf("url %v", itm.url)
			arr, err := utilhttp.GetBytes(itm.url, itm.query, nil)

			if err != nil {
				t.Errorf("Error : %v", err)
			}

			for _, v := range itm.search {
				if !strings.Contains(string(arr), v) {
					t.Errorf("Error on %v", itm.url)
				}
			}

		})

	}

	cmd.Stop()

	time.Sleep(1 * time.Second)

}
