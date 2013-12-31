package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nemothekid/tinybusters/datastore"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	CLOSE_NO_SLOTS     = 4502
	CLOSE_INVALID_PASS = 4503
	CLOSE_INVALID_USER = 4504
)

type GameServer struct {
	conf        ServerConfig
	ServersList []datastore.Server
	NoUsers     int64
	sm          *http.ServeMux
	GameLobby   *Lobby
	datastore   datastore.DataStore
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
		w.WriteHeader(404)
		w.Write(j)
		return
	}
}

func (g *GameServer) getServerInfo() datastore.Server {
	hs := g.conf.HostName
	if hs == "" {
		hs = g.conf.ListenAddress
		if hs == "" || hs == "0.0.0.0" || hs == "0:0:0:0:0:0:0:0" || hs == "::" {
			hs, _ = os.Hostname()
		}
	}
	if hs == "{hostname}" {
		hs, _ = os.Hostname()
	}
	sn := g.conf.ServerName
	if g.conf.ServerName == "" {
		sn = hs
	}
	si := datastore.Server{
		Hostname:  hs,
		Port:      g.conf.GamePort,
		Users:     g.NoUsers,
		Slots:     g.conf.Slots,
		Name:      sn,
		Mode:      g.conf.Mode,
		Updated:   time.Now(),
		ForceAuth: g.conf.ForceAuth,
	}
	return si
}

func (g *GameServer) serverLeaders(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sort := r.FormValue("sort")
	limitS := r.FormValue("limit")
	skipS := r.FormValue("skip")
	spreadS := r.FormValue("spread")
	username := r.FormValue("user")

	var limit int
	var skip int
	var spread int
	var user *datastore.User

	switch sort {
	case datastore.SORT_KILLS:
	case datastore.SORT_DEATHS:
	case datastore.SORT_SCORE:
	case datastore.SORT_WINS:
	case datastore.SORT_LOSSES:
	case datastore.SORT_PLAYS:
	default:
		sort = datastore.SORT_NONE
	}

	if limitS == "" {
		limit = 50
	} else {
		if l, err := strconv.Atoi(limitS); err == nil {
			limit = l
		} else {
			limit = 50
		}
	}

	if skipS == "" {
		skip = 0
	} else {
		if s, err := strconv.Atoi(skipS); err == nil {
			skip = s
		} else {
			skip = 0
		}
	}

	if spreadS == "" {
		spread = 5
	} else {
		if s, err := strconv.Atoi(spreadS); err == nil {
			spread = s
		} else {
			spread = 5
		}
	}

	if username != "" {
		var err error
		user, err = g.datastore.GetUser(username)
		if err == datastore.ErrUserNotFound {
			user = nil
		}
	}
	var leaders []*datastore.User
	var err error
	if user == nil {
		leaders, err = g.datastore.GetUsers(sort, limit, skip)
	} else {
		leaders, err = g.datastore.GetUsersAdjacent(user, sort, spread)
	}

	if err != nil {
		j, _ := json.Marshal(map[string]interface{}{"code": 500, "error": "Server error."})
		w.WriteHeader(404)
		w.Write(j)
		return
	}

	j, _ := json.Marshal(leaders)
	w.Header().Set("Time-Taken", fmt.Sprintf("%v", time.Since(start)))
	w.Write(j)
	return

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

		username := r.FormValue("name")
		if username == "" {
			doE(400, "Invalid name.")
		}
		password := r.FormValue("pass")
		register := r.FormValue("register")

		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			doE(405, "Not a websocket handshake.")
			return
		} else if err != nil {
			log.Println(err)
			doE(500, err.Error())
			return
		}

		user, uerr := g.datastore.GetUser(username)

		if uerr == nil {
			if !user.CheckPassword(password) {
				ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(CLOSE_INVALID_PASS, "Invalid password."), time.Now().Add(10*time.Second))
				return
			}
		} else if uerr == datastore.ErrUserNotFound {
			if g.conf.ForceAuth && password == "" {
				ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(CLOSE_INVALID_PASS, "Invalid password."), time.Now().Add(10*time.Second))
				return
			}
			if password != "" {
				if register != "" {
					user = datastore.NewUser(username, password)
				} else {
					ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(CLOSE_INVALID_USER, "Invalid user."), time.Now().Add(10*time.Second))
					return
				}
			}
			user = datastore.NewUser(username, "")
			user.Temporary = true
		} else {
			ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1011, "Server error."), time.Now().Add(10*time.Second))
			return
		}

		if g.conf.Slots != 0 {
			if atomic.AddInt64(&g.NoUsers, 1) > int64(g.conf.Slots) {
				ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(CLOSE_NO_SLOTS, "No available slots."), time.Now().Add(10*time.Second))
				atomic.AddInt64(&g.NoUsers, -1)
				return
			}
		} else {
			atomic.AddInt64(&g.NoUsers, 1)
		}
		defer atomic.AddInt64(&g.NoUsers, -1)

		player := NewPlayer(ws, user)
		g.GameLobby.Register(player)
		user.Online = true
		g.datastore.PutUser(user)
		if user.Temporary {
			defer g.datastore.DeleteUser(user)
		}
		player.Listen()
		return
	}
	doE(404, "Not found.")
}

func NewGameServer(conf ServerConfig) (*GameServer, error) {
	g := &GameServer{}
	g.conf = conf
	g.ServersList = append(make([]datastore.Server, 0, 4), g.getServerInfo())
	g.sm = http.NewServeMux()
	g.sm.HandleFunc("/info", g.serverInfo)
	g.sm.HandleFunc("/servers", g.serverInfo)
	g.sm.HandleFunc("/connect", g.serverConnect)
	g.sm.HandleFunc("/leaderboard", g.serverLeaders)

	var err error
	switch g.conf.Datastore {
	case datastore.STORE_LEVELDB:
		g.datastore, err = datastore.NewLevelDataStore(g.conf.LevelPath)
		if err != nil {
			return nil, err
		}
	default:
		g.datastore, err = datastore.NewNoneDataStore()
		if err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (g *GameServer) Serve() {
	log.Println("[Server] Starting game server on", fmt.Sprintf("%s:%d", g.conf.HostName, g.conf.GamePort))

	g.GameLobby = NewLobby()
	go g.GameLobby.Listen()
	if e := http.ListenAndServe(fmt.Sprintf("%s:%d", g.conf.ListenAddress, g.conf.GamePort), g.sm); e != nil {
		log.Println("[Server] Failed to start game server on", fmt.Sprintf("%s:%d", g.conf.HostName, g.conf.GamePort), e)
	}
}
