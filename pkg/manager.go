package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
)

type NewLobbyResp struct {
	Name string `json:"name"`
}

// Main has:
// - map from lobby names to lobbies
// - mutex for lobby map
var lobbies map[string]*lobby
var lobbiesLock sync.Mutex
var lobbyDone chan string

// Add lobby to map function:
// - lock mutex, add lobby, unlock mutex
func addLobby(l *lobby) {
	lobbiesLock.Lock()
	lobbies[l.name] = l
	lobbiesLock.Unlock()
}

// Remove lobby from map function:
// - lock mutex, remove lobby, unlock mutex
func removeLobby(name string) {
	lobbiesLock.Lock()
	delete(lobbies, name)
	lobbiesLock.Unlock()
}

// Join lobby connection handler:
// - if the lobby doesn't exist, create it, and add it to map
// - create the user object
// - add the user to the lobby user list
// - start the lobby runner function
// - start the user handler function to make it start listening
func handleJoinLobby(w http.ResponseWriter, r *http.Request) {
	lobbyName := r.URL.Query()["lobby"][0]
	userId := r.URL.Query()["user"][0]

	fmt.Println("Join request from", userId, "for lobby", lobbyName)

	if _, ok := lobbies[lobbyName]; !ok {
		fmt.Println("Creating lobby", lobbyName)
		l := createLobby(lobbyName, lobbyDone)
		addLobby(&l)
		go l.runLobby()
	}
	u := createUser(userId)
	u.runUser(w, r)
	lobbies[lobbyName].addUser(&u)
}

// Lobby cleanup goroutine:
// - selects over the lobby done channel
// - if it hears a done message, remove lobby from map, and keep selecting
func lobbyCleanup() {
	for lobbyName := range lobbyDone {
		fmt.Println("cleaning up", lobbyName)
		removeLobby(lobbyName)
	}
}

func handleValid(w http.ResponseWriter, r *http.Request) {
	var f LobbyForm
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Received validation request:", f)

	// validate the lobby exists
	if _, ok := lobbies[f.Lobby]; !ok {
		http.Error(w, "Invalid lobby", http.StatusBadRequest)
		return
	}

	// implicitly return 200 status ok
}

func randName() string {
	letters := []byte("abcdefghjkmnopqrstuvwxyz")
	rand.Shuffle(len(letters), func(i, j int) {
		letters[i], letters[j] = letters[j], letters[i]
	})
	return string(letters[:4])
}

// Don't actually create a lobby, just generate a lobby code that's 4 random letters
func handleNewLobby(w http.ResponseWriter, r *http.Request) {
	for {
		name := randName()
		if _, ok := lobbies[name]; !ok {
			w.Header().Set("Content-Type", "application/json")
			resp, _ := json.Marshal(NewLobbyResp{Name: name})
			w.Write(resp)
			return
		}
	}
}

// The manager:
// - initializes lobby map
// - initializes lobby done
// - starts the lobby cleanup goroutine
// - sets up handler join lobby
func managerInit() {
	lobbies = make(map[string]*lobby)
	lobbyDone = make(chan string)
	go lobbyCleanup()

	http.HandleFunc("/joinLobby", handleJoinLobby)
	http.HandleFunc("/valid", handleValid)
	http.HandleFunc("/newLobby", handleNewLobby)
}
