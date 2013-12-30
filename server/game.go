package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type serverInfo struct {
	Hostname string `json:hostname`
	Port     int    `json:port`
	Users    int64  `json:users`
	Slots    int    `json:slots`
	Name     string `json:name`
}

type GameServer struct {
	conf        ServerConfig
	ServersList []serverInfo
	NoUsers     int64
	sm          *http.ServeMux
	GameLobby   *Lobby
}

func (g *GameServer) serverInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path == "/info" {
		j, _ := json.Marshal(g.getServerInfo())
		w.Write(j)
		return
	} else if r.URL.Path == "/servers" {
		j, _ := json.Marshal(g.ServersList)
		w.Write(j)
		return
	} else {
		j, _ := json.Marshal(map[string]interface{}{"code": 404, "error": "Not found."})
		w.Write(j)
		return
	}
}

func (g *GameServer) getServerInfo() serverInfo {
	hs := g.conf.HostName
	if hs == "" {
		hs, _ = os.Hostname()
	}
	sn := g.conf.ServerName
	if g.conf.ServerName == "" {
		sn = hs
	}
	si := serverInfo{
		Hostname: hs,
		Port:     g.conf.GamePort,
		Users:    g.NoUsers,
		Slots:    g.conf.Slots,
		Name:     sn,
	}

	return si
}

func (g *GameServer) serverConnect(w http.ResponseWriter, r *http.Request) {
	doE := func(code int, err string) {
		w.Header().Set("Content-Type", "application/json")
		j, _ := json.Marshal(map[string]interface{}{"code": code, "error": err})
		w.WriteHeader(code)
		w.Write(j)
	}
	if r.URL.Path == "/connect" {
		if r.Method != "GET" {
			doE(405, "Method not allowed.")
			return
		}

		if g.conf.Origin != nil {
			found := false
			for _, cl := range g.conf.Origin {
				if strings.ToLower(r.Header.Get("Origin")) == cl {
					found = true
					break
				}
			}
			if !found {
				doE(403, "Origin not allowed")
			}
		}

		pn := r.FormValue("name")
		if pn == "" {
			doE(400, "Invalid name.")
		}

		if g.conf.Slots != 0 {
			// TODO: Mutex
			if g.NoUsers+1 > int64(g.conf.Slots) {
				doE(503, "No available slots.")
				return
			}
		}

		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			doE(405, "Not a websocket handshake.")
			return
		} else if err != nil {
			log.Println(err)
			doE(500, err.Error())
			return
		}

		player := NewPlayer(ws, pn)
		g.GameLobby.Register(player)
		atomic.AddInt64(&g.NoUsers, 1)
		player.Listen()
		atomic.AddInt64(&g.NoUsers, -1)
		return
	}
	doE(404, "Not found.")
}

func NewGameServer(conf ServerConfig) *GameServer {
	g := &GameServer{}
	g.conf = conf
	g.ServersList = append(make([]serverInfo, 0, 4), g.getServerInfo())
	g.sm = http.NewServeMux()
	g.sm.HandleFunc("/info", g.serverInfo)
	g.sm.HandleFunc("/servers", g.serverInfo)
	g.sm.HandleFunc("/connect", g.serverConnect)
	return g
}

func (g *GameServer) Serve() {
	log.Println("[Server] Starting game server on", fmt.Sprintf("%s:%d", g.conf.HostName, g.conf.GamePort))

	g.GameLobby = NewLobby()
	go g.GameLobby.Listen()
	if e := http.ListenAndServe(fmt.Sprintf("%s:%d", g.conf.HostName, g.conf.GamePort), g.sm); e != nil {
		log.Println("[Server] Failed to start game server on", fmt.Sprintf("%s:%d", g.conf.HostName, g.conf.GamePort), e)
	}
}
