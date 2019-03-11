// Package config provides config-data. The most-basic information is to be
// injected at compile-time. Otherwise, this package stops execution with
// an error, when Load is called. The other configuration-options are loaded
// from a config.ini-file, that is located at ConfigPath, when the Load()
// function is called
package config

import (
	"log"
	"strconv"
	"time"

	"github.com/urfave/cli"
	ini "gopkg.in/ini.v1"
)

// ConfigDependant is a string-flag signalizing, that a cli flag depends on
// which config-file is used
const ConfigDependant = "CONFIG-DEPENDANT"

// IsTest is a flag, that should be set when running tests. If set, this package
// won't panic, if the following variables were not injected at compile-time
var IsTest bool

// Version must be injected using the makefile. If its value is set to "develop"
// this server operates in development-mode, if not specified otherwise via the
// cli. The Version is attached to the version command- and request-output
var Version string

// GitCommit must be injected using the makefile. It is attached to the version
// command- and request-output
var GitCommit string

// BuildTime must be injected using the makefile. It is attached to the version
// command- and request-output
var BuildTime string

// ConfigPath must be injected using the makefile. It holds the path to the
// config-file used, if not specified otherwise via the cli
var ConfigPath string

// Server holds the local ip and port and, whether the server runs in production
// or development mode
var Server *ServerConfig

// SSL holds the path to the ssl-certificate
var SSL *SSLConfig

// StreamBase holds the inactivity-timeouts for Streams and StreamSources
var StreamBase *StreamBaseConfig

// ServerConfig holds the local ip and port and, whether the server runs in
// production or development mode
type ServerConfig struct {
	IP   string `ini:"ip"`
	Port int    `ini:"port"`
	// (production = 1; development = 2)
	Mode int `ini:"mode"`
}

// SSLConfig holds the path to the ssl-certificate
type SSLConfig struct {
	CertificatePath string `ini:"path"`
}

// StreamBaseConfig holds the inactivity-timeouts for Streams and StreamSources
type StreamBaseConfig struct {
	SupplierTimeout time.Duration `ini:"suppliertimeout"`
	StreamTimeout   time.Duration `ini:"streamtimeout"`
}

// config is just a wrapper for parsing the ini-file
var config struct {
	SC   ServerConfig     `ini:"server"`
	SSLC SSLConfig        `ini:"ssl"`
	SBC  StreamBaseConfig `ini:"streambase"`
}

// Options returns a list of flags for the cli, which represent the
// configuration-options of the config-file
func Options() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: ConfigPath,
			Usage: "path to configuration-file",
		},
		cli.StringFlag{
			Name:  "server_ip, i",
			Value: ConfigDependant,
			Usage: "ip holds this server's local ip address",
		},
		cli.StringFlag{
			Name:  "server_port, p",
			Value: ConfigDependant,
			Usage: "port holds this server's local port",
		},
		cli.StringFlag{
			Name:  "server_mode, m",
			Value: ConfigDependant,
			Usage: "mode controls the servers bahavior on errors - production (1), development (2) or version-dependant (everything else)",
		},
		cli.StringFlag{
			Name:  "ssl_path, s",
			Value: ConfigDependant,
			Usage: "path holds the path to the ssl-certificate",
		},
		cli.StringFlag{
			Name:  "streambase_suppliertimeout",
			Value: ConfigDependant,
			Usage: "suppliertimeout holds the time in seconds, after which a supplier of character streams is deleted, if inactive",
		},
		cli.StringFlag{
			Name:  "streambase_streamtimeout",
			Value: ConfigDependant,
			Usage: "streamtimeout holds the time in seconds, after which a character stream is closed, no matter its activity",
		},
	}
}

// Load updates that part of the configuration-data, which is derived from the
// config-file located at ConfigPath
func Load(ctx *cli.Context) error {
	if !IsTest && (Version == "" || GitCommit == "" || BuildTime == "" || ConfigPath == "") {
		log.Fatal("corrupted build (the configuration was not injected correctly)")
	}
	if ctx != nil {
		ConfigPath = ctx.String("config")
		f, err := ini.Load(ConfigPath)
		if err != nil {
			log.Fatal("invalid config-path: " + err.Error())
		}

		err = f.MapTo(&config)
		if err != nil {
			log.Fatal("invalid config-format: " + err.Error())
		}

		if ctx.String("server_ip") != ConfigDependant {
			config.SC.IP = ctx.String("server_ip")
		}
		if ctx.String("server_port") != ConfigDependant {
			config.SC.Port, err = strconv.Atoi(ctx.String("server_port"))
			if err != nil {
				log.Fatal("invalid server_port flag")
			}
		}
		if ctx.String("server_mode") != ConfigDependant {
			config.SC.Mode, err = strconv.Atoi(ctx.String("server_mode"))
			if err != nil {
				log.Fatal("invalid server_mode flag")
			}
		}
		if ctx.String("ssl_path") != ConfigDependant {
			config.SSLC.CertificatePath = ctx.String("ssl_path")
		}
		if ctx.String("streambase_suppliertimeout") != ConfigDependant {
			config.SBC.SupplierTimeout, err = time.ParseDuration(ctx.String("streambase_suppliertimeout") + "s")
			if err != nil {
				log.Fatal("invalid streambase_suppliertimeout flag")
			}
		}
		if ctx.String("streambase_streamtimeout") != ConfigDependant {
			config.SBC.StreamTimeout, err = time.ParseDuration(ctx.String("streambase_streamtimeout") + "s")
			if err != nil {
				log.Fatal("invalid streambase_streamtimeout flag")
			}
		}
	}

	config.SC.Mode = evalActualMode(config.SC.Mode)
	Server = &config.SC
	SSL = &config.SSLC
	StreamBase = &config.SBC
	return nil
}

func evalActualMode(climode int) int {
	if climode == 1 {
		return 1
	} else if climode == 2 {
		return 2
	} else {
		if Version == "development" {
			return 1
		} else {
			return 2
		}
	}
}
