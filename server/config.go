package server

import (
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
)

const (
	MODE_DEVELOPMENT = "development"
	MODE_PRODUCTION  = "production"

	STORE_NONE    = "none"
	STORE_LEVELDB = "leveldb"

	GAME_PORT   = 9001
	CLIENT_PORT = 8080
	INFO_PORT   = 8888

	DEF_SLOTS = 0
)

type ServerConfig struct {
	Mode       string `yaml:mode`
	ServerName string `yaml:name`

	HostName   string `yaml:hostname`
	GamePort   int    `yaml:gameport`
	ClientPort int    `yaml:clientport`
	InfoPort   int    `yaml:infoport`

	ClientAssets      string `yaml:clientassets`
	CompiledAssetPath string `yaml:compiledassets`

	CoffeePath string `yaml:coffee`
	LessPath   string `yaml:less`
	UglifyPath string `yaml:uglify`

	Slots int `yaml:slots`

	Origin []string `yaml:origin`

	Datastore string `yaml:datastore`
	LevelPath string `yaml:level_path`

	TmpDir []string

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

		ClientAssets:      "./client",
		CompiledAssetPath: "",

		CoffeePath: "/usr/local/bin/coffee",
		LessPath:   "/usr/local/bin/lessc",
		UglifyPath: "/usr/local/bin/uglifyjs",

		Datastore: STORE_LEVELDB,
		LevelPath: "",

		Slots: DEF_SLOTS,

		TmpDir: make([]string, 0, 2),
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

	makeDirs := func() {
		if sc.Datastore == STORE_LEVELDB {
			if sc.LevelPath == "" {
				var e error
				if sc.LevelPath, e = ioutil.TempDir("", "tblvl"); e != nil {
					sc.Datastore = STORE_NONE
				} else {
					sc.TmpDir = append(sc.TmpDir, sc.LevelPath)
				}
			}
		}

		if sc.CompiledAssetPath == "" {
			var e error
			if sc.CompiledAssetPath, e = ioutil.TempDir("", "tbc"); e != nil {
				sc.CompiledAssetPath = ""
			} else {
				sc.TmpDir = append(sc.TmpDir, sc.CompiledAssetPath)
			}
		}
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("[Init] Failed to open config file %s. Using default settings.", configFile)
		readFlags()
		makeDirs()
		return sc, nil
	}

	if err := goyaml.Unmarshal(data, &sc); err != nil {
		log.Printf("[Init] Failed to parse config file %s.", configFile)
		return sc, err
	}

	readFlags()
	makeDirs()
	return sc, nil
}
