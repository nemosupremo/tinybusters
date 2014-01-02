package server

import (
	"github.com/nemothekid/tinybusters/datastore"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"net"
	"os"
	"strings"
)

const (
	MODE_DEVELOPMENT = "development"
	MODE_PRODUCTION  = "production"

	GAME_PORT   = 9001
	CLIENT_PORT = 8080

	DEF_SLOTS = 0
)

type ServerConfig struct {
	Mode       string `yaml:"mode"`
	ServerName string `yaml:"name"`

	HostName      string `yaml:"hostname"`
	ListenAddress string `yaml:"listen_address"`
	GamePort      int    `yaml:"gameport"`
	ClientPort    int    `yaml:"clientport"`

	ClientAssets      string `yaml:"clientassets"`
	CompiledAssetPath string `yaml:"compiledassets"`

	CoffeePath string `yaml:"coffee"`
	LessPath   string `yaml:"less"`
	UglifyPath string `yaml:"uglify"`

	Slots     int  `yaml:"slots"`
	ForceAuth bool `yaml:"force_auth"`

	Origin []string `yaml:"origin"`

	Datastore     string            `yaml:"datastore"`
	DatastoreConf map[string]string `yaml:"datastore_conf"`

	SeedServers  []string `yaml:"seed_servers"`
	RegisterWith []string `yaml:"register_with"`

	TmpDir []string `yaml:"-"`

	Quit func() `yaml:"-"`
}

func ReadConfig() (ServerConfig, error) {
	configFile := Flags.ConfigFile

	sc := ServerConfig{
		Mode: MODE_DEVELOPMENT,

		ListenAddress: "",
		HostName:      "",
		GamePort:      GAME_PORT,
		ClientPort:    CLIENT_PORT,

		ClientAssets:      "./client",
		CompiledAssetPath: "",
		ForceAuth:         false,

		CoffeePath: "/usr/local/bin/coffee",
		LessPath:   "/usr/local/bin/lessc",
		UglifyPath: "/usr/local/bin/uglifyjs",

		Datastore:     datastore.STORE_LEVELDB,
		DatastoreConf: map[string]string{"level_dbpath": ""},

		Slots: DEF_SLOTS,

		SeedServers:  []string{},
		RegisterWith: []string{},

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

		if Flags.Register != "" {
			sc.RegisterWith = strings.Split(Flags.Register, ",")
		}
	}

	makeDirs := func() {

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

func (sc ServerConfig) GetHostname() string {
	hs := sc.HostName
	if hs == "" {
		hs = sc.ListenAddress
		if hs == "" || hs == net.IPv4zero.String() || hs == "["+net.IPv6zero.String()+"]" || hs == "[::]" {
			hs, _ = os.Hostname()
		}
	}
	if hs == "{hostname}" {
		hs, _ = os.Hostname()
	}
	if hs == "" {
		hs = "localhost"
	}
	return hs
}
