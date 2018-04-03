package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/michael1011/lightningtip/backends"
	"github.com/op/go-logging"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

const (
	defaultConfigFile = "lightningTip.conf"

	defaultDataDir = "LightningTip"

	defaultLogFile  = "lightningTip.log"
	defaultLogLevel = "info"

	defaultRESTHost    = "0.0.0.0:8081"
	defaultTLSCertFile = ""
	defaultTLSKeyFile  = ""

	defaultAccessDomain = ""

	defaultTipExpiry = 3600

	defaultReconnect = false

	defaultLndGRPCHost  = "localhost:10009"
	defaultLndCertFile  = "tls.cert"
	defaultMacaroonFile = "invoice.macaroon"
)

type config struct {
	ConfigFile string `long:"config" Description:"Location of the config file"`

	DataDir string `long:"datadir" Description:"Location of the data stored by LightningTip"`

	LogFile  string `long:"logfile" Description:"Location of the log file"`
	LogLevel string `long:"loglevel" Description:"Log level: debug, info, warning, error"`

	RESTHost    string `long:"resthost" Description:"Host for the REST interface of LightningTip"`
	TLSCertFile string `long:"tlscertfile" Description:"Certificate for using LightningTip via HTTPS"`
	TLSKeyFile  string `long:"tlskeyfile" Description:"Certificate for using LightningTip via HTTPS"`

	AccessDomain string `long:"accessdomain" Description:"The domain you are using LightningTip from"`

	TipExpiry int64 `long:"tipexpiry" Description:"Invoice expiry time in seconds"`

	Reconnect bool `long:"reconnect" Description:"If enabled LightningTip will try to reconnect to LND instead of exiting"`

	LND *backends.LND `group:"LND" namespace:"lnd"`
}

var cfg config

var backend backends.Backend

func initConfig() {
	cfg = config{
		ConfigFile: path.Join(getDefaultDataDir(), defaultConfigFile),

		DataDir: getDefaultDataDir(),

		LogFile:  path.Join(getDefaultDataDir(), defaultLogFile),
		LogLevel: defaultLogLevel,

		RESTHost:    defaultRESTHost,
		TLSCertFile: defaultTLSCertFile,
		TLSKeyFile:  defaultTLSKeyFile,

		AccessDomain: defaultAccessDomain,

		TipExpiry: defaultTipExpiry,

		Reconnect: defaultReconnect,

		LND: &backends.LND{
			GRPCHost:     defaultLndGRPCHost,
			CertFile:     path.Join(getDefaultLndDir(), defaultLndCertFile),
			MacaroonFile: path.Join(getDefaultLndDir(), defaultMacaroonFile),
		},
	}

	// Ignore unknown flags the first time parsing command line flags to prevent showing the unknown flag error twice
	flags.NewParser(&cfg, flags.IgnoreUnknown).Parse()

	errFile := flags.IniParse(cfg.ConfigFile, &cfg)

	// Parse flags again to override config file
	_, err := flags.Parse(&cfg)

	// Default log level if parsing fails
	logLevel := logging.DEBUG

	switch strings.ToLower(cfg.LogLevel) {
	case "info":
		logLevel = logging.INFO

	case "warning":
		logLevel = logging.WARNING

	case "error":
		logLevel = logging.ERROR
	}

	// Create data directory
	var errDataDir error
	var dataDirCreated bool

	if _, err := os.Stat(getDefaultDataDir()); os.IsNotExist(err) {
		errDataDir = os.Mkdir(getDefaultDataDir(), 0700)

		dataDirCreated = true
	}

	errLogFile := initLogger(cfg.LogFile, logLevel)

	// Show error messages
	if err != nil {
		log.Error("Failed to parse command line flags")
	}

	if errDataDir != nil {
		log.Error("Could not create data directory")
		log.Debug("Data directory path: " + getDefaultDataDir())

	} else if dataDirCreated {
		log.Debug("Created data directory: " + getDefaultDataDir())
	}

	if errFile != nil {
		log.Warning("Failed to parse config file: " + fmt.Sprint(errFile))
	} else {
		log.Debug("Parsed config file: " + cfg.ConfigFile)
	}

	if errLogFile != nil {
		log.Error("Failed to initialize log file: " + fmt.Sprint(err))

	} else {
		log.Debug("Initialized log file: " + cfg.LogFile)
	}

	backends.UseLogger(*log)

	backend = cfg.LND
}

func getDefaultDataDir() (dir string) {
	homeDir := getHomeDir()

	switch runtime.GOOS {
	case "windows":
		fallthrough

	case "darwin":
		dir = path.Join(homeDir, defaultDataDir)

	default:
		dir = path.Join(homeDir, "."+strings.ToLower(defaultDataDir))
	}

	return dir
}

func getDefaultLndDir() (dir string) {
	homeDir := getHomeDir()

	switch runtime.GOOS {
	case "darwin":
		fallthrough

	case "windows":
		dir = path.Join(homeDir, "Lnd")

	default:
		dir = path.Join(homeDir, ".lnd")
	}

	return dir
}

func getHomeDir() (dir string) {
	usr, err := user.Current()

	if err == nil {
		switch runtime.GOOS {
		case "darwin":
			dir = path.Join(usr.HomeDir, "Library/Application Support")

		case "windows":
			dir = path.Join(usr.HomeDir, "AppData/Local")

		default:
			dir = usr.HomeDir
		}

	}

	return dir
}
