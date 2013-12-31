package datastore

import (
	"bytes"
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"fmt"
	"time"
)

const (
	STORE_NONE    = "none"
	STORE_LEVELDB = "leveldb"

	SORT_NONE   = "none"
	SORT_KILLS  = "kills"
	SORT_DEATHS = "deaths"
	SORT_SCORE  = "score"
	SORT_WINS   = "wins"
	SORT_LOSSES = "losses"
	SORT_PLAYS  = "plays"

	idLength = 64
)

var ErrUserNotFound = fmt.Errorf("User not found.")
var ErrServerNotFound = fmt.Errorf("Server not found.")

type DataStore interface {
	GetUsers(string, int, int) ([]*User, error)
	GetUsersAdjacent(*User, string, int) ([]*User, error)
	GetUser(string) (*User, error)
	PutUser(*User) error
	DeleteUser(*User) error
	NumUsers() (int, error)

	GetServers() ([]*Server, error)
	GetServer([]byte) (*Server, error)
	PutServer(*Server) error
	DeleteServer(*Server) error
	NumServers() (int, error)

	Close()
}

type User struct {
	Id       []byte `msgpack:"id" json:"-"`
	Name     string `msgpack:"n" json:"name"`
	Password []byte `msgpack:"p" json:"-"`

	Online bool `msgpack:"o" json:"online"`

	Kills  int64 `msgpack:"k" json:"kills"`
	Deaths int64 `msgpack:"d" json:"deaths"`
	Score  int64 `msgpack:"s" json:"score"`

	Wins   int64 `msgpack:"w" json:"wins"`
	Losses int64 `msgpack:"l" json:"losses"`
	Plays  int64 `msgpack:"pl" json:"games_played"`

	Temporary bool `msgpack:"tmp" json:"temporary"`
}

type Server struct {
	Hostname  string    `json:"hostname" msgpack:"h"`
	Port      int       `json:"port" msgpack:"p"`
	Users     int64     `json:"users" msgpack:"u"`
	Slots     int       `json:"slots" msgpack:"s"`
	Name      string    `json:"name" msgpack:"n"`
	Mode      string    `json:"mode" msgpack:"m"`
	ForceAuth bool      `json:"force_auth" msgpack:"fa"`
	Updated   time.Time `json:"updated" msgpack:"up"`
}

func NewUser(username, password string) *User {
	user := &User{
		Kills:  0,
		Deaths: 0,
		Score:  0,
		Wins:   0,
		Losses: 0,
		Plays:  0,

		Online:    false,
		Temporary: false,
	}
	user.Id = make([]byte, idLength)
	if _, err := rand.Read(user.Id); err != nil {
		panic("Failed to read random value into salt. (" + err.Error() + ")")
	}
	user.Name = username
	if password != "" {
		user.Password, _ = bcrypt.GenerateFromPassword(bytes.Join([][]byte{user.Id, []byte(password)}, []byte(".")), bcrypt.MinCost)
	} else {
		user.Password = nil
	}
	return user
}

func (u *User) CheckPassword(password string) bool {
	if u.Password != nil {
		return bcrypt.CompareHashAndPassword(u.Password, bytes.Join([][]byte{u.Id, []byte(password)}, []byte("."))) == nil
	}
	return true
}
