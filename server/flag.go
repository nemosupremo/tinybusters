package server

import (
	"flag"
)

var Flags *FlagConfig

type FlagConfig struct {
	Mode string

	HostName   string
	GamePort   int
	ClientPort int
	InfoPort   int

	ConfigFile string

	Register string
}

func init() {
	Flags = flags()
}

func flags() *FlagConfig {
	fc := new(FlagConfig)
	//hostname, _ := os.Hostname()
	flag.StringVar(&(fc.ConfigFile), "config", "./tb.yaml", "Config file location.")
	flag.StringVar(&(fc.Mode), "mode", "", "Server mode (development or production).")
	flag.StringVar(&(fc.HostName), "hostname", "", "Hostname to listen on.")
	flag.StringVar(&(fc.Register), "register", "", "Register this server with another server (hostname:port).")
	flag.IntVar(&(fc.GamePort), "gameport", GAME_PORT, "Game Server Port")
	flag.IntVar(&(fc.ClientPort), "clientport", CLIENT_PORT, "HTTP Client Port (0 for disabled)")
	return fc

}
