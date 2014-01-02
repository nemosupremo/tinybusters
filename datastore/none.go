package datastore

const (
	STORE_NONE = "none"
)

func init() {
	RegisterStore(STORE_NONE, NewLevelDataStore)
}

type NoneDataStore struct{}

func NewNoneDataStore(conf map[string]string) (DataStore, error) {
	return new(NoneDataStore), nil
}

func (n *NoneDataStore) GetUsers(string, int, int) ([]*User, error) {
	return []*User{}, nil
}

func (n *NoneDataStore) GetUsersAdjacent(*User, string, int) ([]*User, error) {
	return []*User{}, nil
}

func (n *NoneDataStore) GetUser(string) (*User, error) {
	return nil, ErrUserNotFound
}

func (n *NoneDataStore) PutUser(*User) error {
	return nil
}

func (n *NoneDataStore) DeleteUser(*User) error {
	return nil
}

func (n *NoneDataStore) NumUsers() (int, error) {
	return 0, nil
}

func (n *NoneDataStore) GetServers() ([]*Server, error) {
	return []*Server{}, nil
}
func (n *NoneDataStore) GetServer(string) (*Server, error) {
	return nil, ErrServerNotFound
}
func (n *NoneDataStore) PutServer(*Server) error {
	return nil
}
func (n *NoneDataStore) DeleteServer(*Server) error {
	return nil
}

func (n *NoneDataStore) NumServers() (int, error) {
	return 0, nil
}

func (n *NoneDataStore) GetFriends(*User) ([]*User, error) {
	return []*User{}, nil
}

func (n *NoneDataStore) AddFriend(*User, *User) error {
	return nil
}

func (n *NoneDataStore) RemoveFriend(*User, *User) error {
	return nil
}

func (n *NoneDataStore) Close() {
}
