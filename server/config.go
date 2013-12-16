package server

import (
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
)

const (
	MODE_DEVELOPMENT = "development"
	MODE_PRODUCTION  = "production"

	GAME_PORT   = 9001
	CLIENT_PORT = 8080
	INFO_PORT   = 8888
)

type ServerConfig struct {
	Mode string `yaml:mode`

	HostName   string `yaml:hostname`
	GamePort   int    `yaml:gameport`
	ClientPort int    `yaml:clientport`
	InfoPort   int    `yaml:infoport`

	ClientAssets string `yaml:clientassets`

	CoffeePath string `yaml:coffee`
	LessPath   string `yaml:less`

	Quit func()
}

func ReadConfig() (ServerConfig, error) {
	configFile := Flags.ConfigFile

	sc := ServerConfig{
		Mode:       MODE_DEVELOPMENT,
		HostName:   "",
		GamePort:   GAME_PORT,
		ClientPort: CLIENT_PORT,
		InfoPort:   INFO_PORT,

		ClientAssets: "./client",

		CoffeePath: "/usr/local/bin/coffee",
		LessPath:   "/usr/bin/less",
	}

	readFlags := func() {
		if Flags.Mode != "" {
			switch Flags.Mode {
			case MODE_DEVELOPMENT, MODE_PRODUCTION:
				sc.Mode = Flags.Mode
			}
		}
		if Flags.HostName != "" {
			sc.HostName = Flags.HostName
		}
		if Flags.GamePort != GAME_PORT {
			sc.GamePort = Flags.GamePort
		}
		if Flags.ClientPort != CLIENT_PORT {
			sc.ClientPort = Flags.ClientPort
		}
		if Flags.InfoPort != INFO_PORT {
			sc.InfoPort = Flags.InfoPort
		}

	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("[Init] Failed to open config file %s. Using default settings.", configFile)
		readFlags()
		return sc, nil
	}

	if err := goyaml.Unmarshal(data, &sc); err != nil {
		log.Printf("[Init] Failed to parse config file %s.", configFile)
		return sc, err
	}

	readFlags()
	return sc, nil
}
