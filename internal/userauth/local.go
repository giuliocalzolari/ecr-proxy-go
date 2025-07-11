package userauth

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewLocalAuth loads users from a JSON file.
// The JSON file should be an array of objects: [{"username":"user1","password":"pass1"}, ...]
type LocalAuth struct {
	users    map[string]string
	mu       sync.RWMutex
	jsonPath string
	quit     chan struct{}
}

// NewLocalAuth loads users from a JSON file and starts a goroutine to periodically reload it.
func NewLocalAuth(jsonPath string) (*LocalAuth, error) {
	la := &LocalAuth{
		users:    make(map[string]string),
		jsonPath: jsonPath,
		quit:     make(chan struct{}),
	}

	if err := la.reload(); err != nil {
		return nil, err
	}

	go la.periodicReload(30 * time.Second) // reload every 30 seconds

	return la, nil
}

// reload loads the users from the JSON file.
func (la *LocalAuth) reload() error {
	file, err := os.Open(la.jsonPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var userList []User
	if err := json.NewDecoder(file).Decode(&userList); err != nil {
		return err
	}

	users := make(map[string]string)
	for _, u := range userList {
		users[u.Username] = u.Password
	}

	la.mu.Lock()
	la.users = users
	la.mu.Unlock()
	return nil
}

// periodicReload reloads the user database at the given interval.
func (la *LocalAuth) periodicReload(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			la.reload()
		case <-la.quit:
			return
		}
	}
}

// Close stops the periodic reload goroutine.
func (la *LocalAuth) Close() {
	close(la.quit)
}

// IsValid checks if the provided username and password are valid.
func (la *LocalAuth) IsValid(username, password string) bool {
	la.mu.RLock()
	pass, ok := la.users[username]
	la.mu.RUnlock()
	return ok && pass == password
}
