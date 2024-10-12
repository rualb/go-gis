// Package config app config
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-gis/internal/config/consts"
	"go-gis/internal/tool/toolconfig"
	"go-gis/internal/tool/toolhttp"
	xlog "go-gis/internal/tool/toollog"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

var (
	AppVersion  = ""
	AppCommit   = ""
	AppDate     = ""
	ShortCommit = ""
)

func dumpVersionAndExitIf() {

	if CmdLine.Version {
		fmt.Printf("Version: %s\n", AppVersion)
		fmt.Printf("Commit: %s\n", AppCommit)
		fmt.Printf("Date: %s\n", AppDate)
		//
		os.Exit(0)
	}

}

type CmdLineConfig struct {
	Config     string
	CertDir    string
	ConfigsDir string
	Env        string
	Name       string
	Version    bool

	SysAPIKey string
	Listen    string
	ListenTLS string
	ListenSys string

	DumpConfig bool
}

const (
	envDevelopment = "development"
	envTesting     = "testing"
	envStaging     = "staging"
	envProduction  = "production"
)

var envNames = []string{
	envDevelopment, envTesting, envStaging, envProduction,
}

var CmdLine = CmdLineConfig{}

// ReadFlags read app flags
func ReadFlags() {

	_ = os.Args
	flag.StringVar(&CmdLine.Config, "config", "", "Path to dir with config files")
	flag.StringVar(&CmdLine.CertDir, "cert-dir", "", "Path to dir with cert files")
	flag.StringVar(&CmdLine.SysAPIKey, "sys-api-key", "", "Sys api key")
	flag.StringVar(&CmdLine.Listen, "listen", "", "Listen")
	flag.StringVar(&CmdLine.ListenTLS, "listen-tls", "", "Listen TLS")
	flag.StringVar(&CmdLine.ListenSys, "listen-sys", "", "Listen sys")
	flag.StringVar(&CmdLine.Env, "env", "", "Environment: development, testing, staging, production")
	flag.StringVar(&CmdLine.Name, "name", "", "App name")
	flag.StringVar(&CmdLine.ConfigsDir, "configs-dir", "", "Path to dir with configs")

	flag.BoolVar(&CmdLine.Version, "version", false, "App version")

	flag.BoolVar(&CmdLine.DumpConfig, "dump-config", false, "Dump Config")

	flag.Parse() // dont use from init()

	dumpVersionAndExitIf()

}

type envReader struct {
	envError error
	prefix   string
}

func NewEnvReader() envReader {
	return envReader{prefix: "app_"}
}
func (x *envReader) String(p *string, name string, cmdValue *string) {
	envName := strings.ToUpper(x.prefix + name) // *nix case-sensitive
	if cmdValue != nil && *cmdValue != "" {
		xlog.Info("Reading %q value from cmd: %v", name, *cmdValue)
		*p = *cmdValue
		return
	}

	if envName != "" {
		envValue := os.Getenv(envName)
		if envValue != "" {
			xlog.Info("Reading %q value from env: %v = %v", name, envName, envValue)
			*p = envValue
			return
		}
	}

}

func (x *envReader) Bool(p *bool, name string, cmdValue *bool) {

	envName := strings.ToUpper(x.prefix + name) // *nix case-sensitive

	if cmdValue != nil && *cmdValue {
		xlog.Info("Reading %q value from cmd: %v", name, *cmdValue)
		*p = *cmdValue
		return
	}
	if envName != "" {
		envValue := os.Getenv(envName)
		if envValue != "" {
			xlog.Info("Reading %q value from env: %v = %v", name, envName, envValue)
			*p = envValue == "1" || envValue == "true"
			return
		}
	}
}

func (x *envReader) Float64(p *float64, name string, cmdValue *float64) {

	envName := strings.ToUpper(x.prefix + name) // *nix case-sensitive

	if cmdValue != nil && math.Abs(*cmdValue) > 0.000001 {
		xlog.Info("Reading float64 %q value from cmd: %v", name, *cmdValue)
		*p = *cmdValue
		return
	}

	if envName != "" {
		envValue := os.Getenv(envName)
		if envValue != "" {
			xlog.Info("Reading float64 %q value from env: %v = %v", name, envName, envValue)

			if v, err := strconv.ParseFloat(envValue, 64); err == nil {
				*p = v
			} else {
				x.envError = err
			}

		}
	}

}
func (x *envReader) Int(p *int, name string, cmdValue *int) {

	envName := strings.ToUpper(x.prefix + name) // *nix case-sensitive

	if cmdValue != nil && *cmdValue != 0 {
		xlog.Info("Reading %q value from cmd: %v", name, *cmdValue)
		*p = *cmdValue
		return
	}
	if envName != "" {
		envValue := os.Getenv(envName)
		if envValue != "" {
			xlog.Info("Reading %q value from env: %v = %v", name, envName, envValue)

			if v, err := strconv.Atoi(envValue); err == nil {
				*p = v
			} else {
				x.envError = err
			}

		}
	}

}

type Database struct {
	Dialect   string `json:"dialect"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	Name      string `json:"name"`
	User      string `json:"user"`
	Password  string `json:"password"`
	MaxOpen   int    `json:"max_open"`
	MaxIdle   int    `json:"max_idle"`
	IdleTime  int    `json:"idle_time"`
	Migration bool   `json:"migration"`
}

// type AppConfigLog struct {
// 	Level int `json:"level"` // 0=Error 1=Warn 2=Info 3=Debug
// }

type AppConfigMapsGateway struct {
	Enabled bool   `json:"enabled"`
	APIKey  string `json:"api_key"`
	URL     string `json:"url"`
}

type AppConfigVault struct {
	VaultAuth map[string]string `json:"auth"` // keyId:keyValue
}

type AppConfigLang struct {
	Langs []string `json:"langs"`
}

type AppConfigMod struct {
	Name  string `json:"-"`
	Env   string `json:"env"` // prod||'' dev stage
	Debug bool   `json:"-"`
	Title string `json:"title"`

	ConfigPath []string `json:"-"` // []string{".", os.Getenv("APP_CONFIG"), flagAppConfig}
}
type AppConfig struct {
	AppConfigMod

	// Log AppConfigLog `json:"logger"`

	Vault AppConfigVault `json:"vault"`

	DB    Database `json:"database"`
	Redis Database `json:"redis"`

	Lang AppConfigLang `json:"lang"`

	OsmGateway   AppConfigMapsGateway `json:"osm_gateway"`
	GmapsGateway AppConfigMapsGateway `json:"gmaps_gateway"`
	// gms
	HTTPTransport AppConfigHTTPTransport `json:"http_transport"`

	HTTPServer AppConfigHTTPServer `json:"http_server"`

	Configs AppConfigConfigs `json:"configs"`
}

func NewAppConfig() *AppConfig {

	res := &AppConfig{

		Lang: AppConfigLang{Langs: []string{"en"}},
		// Log: AppConfigLog{
		// 	Level: consts.LogLevelWarn,
		// },

		DB: Database{
			Dialect:  "postgres",
			Host:     "localhost",
			Port:     "5432",
			Name:     "postgres",
			User:     "postgres",
			Password: "postgres",
			MaxOpen:  0,
			MaxIdle:  0,
			IdleTime: 0,
		},

		Redis: Database{
			Host:     "localhost",
			Port:     "6379",
			Name:     "redis",
			User:     "redis",
			Password: "redis",
		},

		AppConfigMod: AppConfigMod{
			Name:       consts.AppName,
			ConfigPath: []string{},
			Title:      "",
			Env:        "production",
			Debug:      false,
		},

		OsmGateway: AppConfigMapsGateway{
			URL: "https://nominatim.openstreetmap.org/search?q={LatLng}&format=json&accept-language={Lang}",
		},

		GmapsGateway: AppConfigMapsGateway{
			URL: "https://maps.googleapis.com/maps/api/geocode/json?latlng={LatLng}&key={Key}&language={Lang}&location_type=ROOFTOP&result_type=street_address",
		},

		HTTPTransport: AppConfigHTTPTransport{},

		HTTPServer: AppConfigHTTPServer{
			ReadTimeout:  0,
			WriteTimeout: 0,
			IdleTimeout:  0,

			RateLimit: 0,
			RateBurst: 0,

			Listen: "localhost:31180",
			//ListenTLS: "localhost:31183",
			CertDir: "",

			SysAPIKey: "",
		},

		Configs: AppConfigConfigs{
			Dir: "",
		},
	}

	return res
}

func (x *AppConfig) readEnvName() error {
	reader := NewEnvReader()
	// APP_ENV -env
	reader.String(&x.Env, "env", &CmdLine.Env)
	reader.String(&x.Name, "name", &CmdLine.Name)

	if err := x.validateEnv(); err != nil {
		return err
	}

	configPath := slices.Concat(strings.Split(os.Getenv("APP_CONFIG"), ";"), strings.Split(CmdLine.Config, ";"))
	configPath = slices.Compact(configPath)
	configPath = slices.DeleteFunc(
		configPath,
		func(x string) bool {
			return x == ""
		},
	)

	for i := 0; i < len(configPath); i++ {
		configPath[i] += "/" + x.Name
	}

	// if len(configPath) == 0 {
	// 	configPath = []string{"."} // default
	// }

	if len(configPath) == 0 {
		xlog.Warn("Config path is empty")
	} else {
		xlog.Info("Config path: %v", configPath)
	}

	x.ConfigPath = configPath

	return nil
}

func (x *AppConfig) readEnvVar() error {
	reader := NewEnvReader()

	// OsmGateway configuration
	reader.String(&x.OsmGateway.URL, "osm_url", nil)
	reader.String(&x.OsmGateway.APIKey, "osm_api_key", nil)
	reader.Bool(&x.OsmGateway.Enabled, "osm_enabled", nil)
	// GoogleMapsGateway configuration
	reader.String(&x.GmapsGateway.URL, "gmaps_url", nil)
	reader.String(&x.GmapsGateway.APIKey, "gmaps_api_key", nil)
	reader.Bool(&x.GmapsGateway.Enabled, "gmaps_enabled", nil)

	// Database configuration

	reader.String(&x.DB.Dialect, "db_dialect", nil)
	reader.String(&x.DB.Host, "db_host", nil)
	reader.String(&x.DB.Port, "db_port", nil)
	reader.String(&x.DB.Name, "db_name", nil)
	reader.String(&x.DB.User, "db_user", nil)
	reader.String(&x.DB.Password, "db_password", nil)
	reader.Int(&x.DB.MaxOpen, "db_max_open", nil)
	reader.Int(&x.DB.MaxIdle, "db_max_idle", nil)
	reader.Int(&x.DB.IdleTime, "db_idle_time", nil)
	reader.Bool(&x.DB.Migration, "db_migration", nil)

	// General configuration
	reader.String(&x.Title, "title", nil)

	// Http server
	reader.Bool(&x.HTTPServer.AccessLog, "http_access_log", nil)
	reader.Float64(&x.HTTPServer.RateLimit, "http_rate_limit", nil)
	reader.Int(&x.HTTPServer.RateBurst, "http_rate_burst", nil)
	reader.String(&x.HTTPServer.Listen, "http_listen", nil)        // =>listen
	reader.String(&x.HTTPServer.ListenTLS, "http_listen_tls", nil) // =>listen_tls
	reader.Bool(&x.HTTPServer.AutoTLS, "http_auto_tls", nil)
	reader.Bool(&x.HTTPServer.RedirectHTTPS, "http_redirect_https", nil)
	reader.Bool(&x.HTTPServer.RedirectWWW, "http_redirect_www", nil)
	reader.String(&x.HTTPServer.CertDir, "http_cert_dir", &CmdLine.CertDir) // =>cert_dir
	reader.Int(&x.HTTPServer.ReadTimeout, "http_read_timeout", nil)
	reader.Int(&x.HTTPServer.WriteTimeout, "http_write_timeout", nil)
	reader.Int(&x.HTTPServer.IdleTimeout, "http_idle_timeout", nil)
	reader.Int(&x.HTTPServer.ReadHeaderTimeout, "http_read_header_timeout", nil)
	reader.String(&x.HTTPServer.ListenSys, "http_listen_sys", nil)  // =>listen_sys
	reader.String(&x.HTTPServer.SysAPIKey, "http_sys_api_key", nil) // =>sys_api_key

	reader.String(&x.HTTPServer.CertDir, "cert_dir", &CmdLine.CertDir) // short
	reader.String(&x.Configs.Dir, "configs_dir", &CmdLine.ConfigsDir)

	reader.String(&x.HTTPServer.Listen, "listen", &CmdLine.Listen)
	reader.String(&x.HTTPServer.ListenTLS, "listen_tls", &CmdLine.ListenTLS)
	reader.String(&x.HTTPServer.ListenSys, "listen_sys", &CmdLine.ListenSys)

	reader.String(&x.HTTPServer.SysAPIKey, "sys_api_key", &CmdLine.SysAPIKey)

	if reader.envError != nil {
		return reader.envError
	}

	return nil
}

func (x *AppConfig) validateEnv() error {

	if x.Env == "" {
		x.Env = envProduction
	}

	x.Debug = x.Env == envDevelopment
	if !slices.Contains(envNames, x.Env) {
		xlog.Warn("Non-standart env name: %v", x.Env)
	}

	return nil

}
func (x AppConfig) validate() error {

	if x.HTTPServer.Listen == "" && x.HTTPServer.ListenTLS == "" {
		return fmt.Errorf("socket Listen and ListenTLS are empty")
	}

	return nil
}

type AppConfigSource struct {
	config *AppConfig
}

func MustNewAppConfigSource() *AppConfigSource {

	res := &AppConfigSource{}

	err := res.Load() // init-load

	if err != nil {
		panic(err)
	}

	return res

}

type AppConfigHTTPTransport struct {
	MaxIdleConns        int `json:"max_idle_conns,omitempty"`
	MaxIdleConnsPerHost int `json:"max_idle_conns_per_host,omitempty"`
	IdleConnTimeout     int `json:"idle_conn_timeout,omitempty"`
	MaxConnsPerHost     int `json:"max_conns_per_host,omitempty"`
}

type AppConfigHTTPServer struct {
	AccessLog     bool    `json:"access_log"`
	RateLimit     float64 `json:"rate_limit"`
	RateBurst     int     `json:"rate_burst"`
	Listen        string  `json:"listen"`
	ListenTLS     string  `json:"listen_tls"`
	AutoTLS       bool    `json:"auto_tls"`
	RedirectHTTPS bool    `json:"redirect_https"`
	RedirectWWW   bool    `json:"redirect_www"`

	CertDir string `json:"cert_dir"`

	ReadTimeout       int `json:"read_timeout,omitempty"`        // 5 to 30 seconds
	WriteTimeout      int `json:"write_timeout,omitempty"`       // 10 to 30 seconds, WriteTimeout > ReadTimeout
	IdleTimeout       int `json:"idle_timeout,omitempty"`        // 60 to 120 seconds
	ReadHeaderTimeout int `json:"read_header_timeout,omitempty"` // default get from ReadTimeout

	SysMetrics bool   `json:"sys_metrics"` //
	SysAPIKey  string `json:"sys_api_key"`
	ListenSys  string `json:"listen_sys"`
}
type AppConfigConfigs struct {
	Dir string `json:"dir"`
}

// Load load config
func (x *AppConfigSource) Load() error {

	res := NewAppConfig()

	{
		err := res.readEnvName()
		if err != nil {
			return err
		}
	}

	{
		for i := 0; i < len(res.ConfigPath); i++ {

			dir := res.ConfigPath[i]

			fileName := fmt.Sprintf("config.%s.json", res.Env)

			xlog.Info("Loading config from: %v", dir)

			err := toolconfig.LoadConfig(res /*pointer*/, dir, fileName)

			if err != nil {
				return err
			}

		}

	}

	{
		err := res.readEnvVar()
		if err != nil {
			return err
		}

	}

	{
		err := res.validate()
		if err != nil {
			return err
		}
	}

	xlog.Info("Config loaded: Name=%v Env=%v Debug=%v ", res.Name, res.Env, res.Debug)

	x.config = res

	if CmdLine.DumpConfig {
		data, _ := json.MarshalIndent(res, "", " ")
		fmt.Println(string(data))
	}

	return nil
}

func (x *AppConfigSource) Config() *AppConfig {

	return x.config

}

// func (x *AppConfig) ApplyConfigFromFilesList(files string, errIfNotExists bool) error {

// 	if files == "" {
// 		return nil
// 	}

// 	for _, x := range strings.Split(files, ";") {
// 		err := x.ApplyConfigFromFile(x, errIfNotExists)

// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// FromFile errIfNotExists argument soft binding, no error if file not exists
func (x *AppConfig) FromFile(dir string, file string) error {

	if file == "" {
		return nil
	}

	if !strings.HasSuffix(file, ".json") && !strings.HasPrefix(file, "config.") {
		return fmt.Errorf("error file not match config.*.json: %v", file)
	}

	fullPath, err := filepath.Abs(filepath.Join(dir, file))

	if err != nil {
		return err
	}

	//
	fullPath = filepath.Clean(fullPath)
	data, err := os.ReadFile(fullPath)

	if err != nil {
		return fmt.Errorf("error with file %v: %v", fullPath, err)
	}

	xlog.Info("Loading config from file: %v", fullPath)

	err = x.FromJSON(string(data))
	if err != nil {
		return err
	}

	return nil
}

// FromURL errIfNotExists argument soft binding, no error if file not exists
func (x *AppConfig) FromURL(dir string, file string) error {

	if file == "" {
		return nil
	}

	if !strings.HasSuffix(file, ".json") && !strings.HasPrefix(file, "config.") {
		return fmt.Errorf("error file not match config.*.json: %v", file)
	}

	fullPath := dir + "/" + file

	_, err := url.Parse(fullPath)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// fmt.Println("Reading config from file: ", file)

	data, err := toolhttp.GetBytes(fullPath, nil, nil)

	if err != nil {
		return fmt.Errorf("error with file %v: %v", fullPath, err)
	}

	xlog.Info("Loading config from file: %v", fullPath)

	err = x.FromJSON(string(data))
	if err != nil {
		return err
	}

	return nil
}

// func (appConfig *AppConfig) ApplyConfigFromEnv(env string) {

// 	appConfig.ApplyConfigFromFile(fmt.Sprintf("config.%s.json", env))

// }

// FromJSON from json
func (x *AppConfig) FromJSON(data string) error {

	if data == "" {
		return nil
	}

	err := json.Unmarshal([]byte(data), x)

	if err != nil {
		return err
	}

	return nil
}
