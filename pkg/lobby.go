package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

const maxLobbySize = 2

// target: either lobby or game
// cmd: specific command, e.g. 'kick', 'changeHost', 'move'
// args: details for what to do with the command
type userMsg struct {
	Target string
	Cmd    string
	Args   map[string]string
}

type lobbyMsg struct {
	Target string
	State  interface{}
}

type lobbyState struct {
	Status  string
	Players []string
	Scores  map[string]int
}

// A lobby has:
// - status (playing or lobby)
// - game object pointer
// - map of user ids to user objects
// - mutex for protecting list of users
// - done channel for signalling main to destroy lobby
// - webToLobby channel for passing to users
type lobby struct {
	name        string
	state       lobbyState
	g           game
	users       map[string]*user
	userLock    sync.Mutex
	done        chan string
	webToLobby  chan []byte
	lobbyToGame chan userMsg
	gameToLobby chan gameMsg
	userEndConn chan string
}

func (l *lobby) broadcastState() {
	fmt.Println("Broadcasting state")
	msg, _ := json.Marshal(lobbyMsg{"lobby", l.state})
	l.userLock.Lock()
	for u := range l.users {
		l.users[u].lobbyToWeb <- msg
	}
	l.userLock.Unlock()
}

// Add user function:
// - set user's webToLobby channel
// - lock mutex
// - add user to list
// - release mutex
func (l *lobby) addUser(u *user) {
	// if the user id is already in the lobby it's probably a reload, so clear the old connection
	if _, ok := l.users[u.id]; ok {
		l.removeUser(u.id)
	}
	u.webToLobby = l.webToLobby
	u.endConn = l.userEndConn
	l.userLock.Lock()
	l.users[u.id] = u
	l.state.Players = append(l.state.Players, u.id)
	l.state.Scores[u.id] = 0
	l.userLock.Unlock()
	l.broadcastState()
	if l.g != nil {
		go l.g.broadcastState()
	}
}

// Remove user function:
// - lock mutex
// - remove user from list
// - release mutex
func (l *lobby) removeUser(id string) {
	l.userLock.Lock()
	idx := 0
	for i, player := range l.state.Players {
		if player == id {
			idx = i
		}
	}
	l.state.Players = append(l.state.Players[:idx], l.state.Players[idx+1:]...)
	delete(l.state.Scores, id)
	delete(l.users, id)
	l.userLock.Unlock()
	l.broadcastState()
}

// Lobby creation involves:
// - setting name
// - setting status
// - creating list of users
// - initializing mutex to protect user list
func createLobby(name string, done chan string) lobby {
	return lobby{
		name: name,
		state: lobbyState{
			Status:  "lobby",
			Players: make([]string, 0),
			Scores:  make(map[string]int),
		},
		users:       make(map[string]*user),
		done:        done,
		webToLobby:  make(chan []byte),
		gameToLobby: make(chan gameMsg),
		userEndConn: make(chan string),
	}
}

func (l *lobby) handleStartGame() {
	fmt.Println("lobby starting game")
	l.gameToLobby = make(chan gameMsg)
	l.lobbyToGame = make(chan userMsg)
	l.g = createBunga(&l.state, l.lobbyToGame, l.gameToLobby)
	l.state.Status = "game"
	fmt.Println("lobby running game")
	go l.g.runGame()
}

func (l *lobby) handleQuitGame() {
	// add scores to lobby state
	bunga, _ := l.g.(*bunga)
	if bunga != nil {
		for player, score := range bunga.state.Scores {
			l.state.Scores[player] += score
		}
	}
	// sort players by scores, ascending
	sort.Slice(l.state.Players, func(i, j int) bool {
		return l.state.Scores[l.state.Players[i]] < l.state.Scores[l.state.Players[j]]
	})
	l.g = nil
	l.lobbyToGame = nil
	l.gameToLobby = nil
}

// TODO: handle error cases here
func (l *lobby) handleCommand(msg *userMsg) {
	fmt.Println("Got gommand:", msg)
	fmt.Println("Target:", msg.Target, "Cmd:", msg.Cmd, "Args:", msg.Args)
	switch msg.Cmd {
	case "startGame":
		l.handleStartGame()

	case "quitGame":
		l.lobbyToGame <- userMsg{"quit", "", nil}
		l.handleQuitGame()

	case "backToLobby":
		l.state.Status = "lobby"

	}

	fmt.Println("Finished running command, broadcasting state")
	l.broadcastState()
}

func (l *lobby) endLobby() {
	l.done <- l.name
	l.lobbyToGame <- userMsg{"game", "quit", nil}
}

// The main lobby routine:
// - starts a watchdog timer for exiting if there's no users
// - selects across all user 'webToUser' channels and watchdog timer
// - if a user is done, remove them from the user list
//   - also send a message to their endConnection channel
// - if a user is done and they're the host, update the host
// - if a user is done and they're playing a game, pass the info to the game
// - if last user is done, return
// - user starts game:
//   - get game initialization function
//   - initialize game
//   - set game object pointer
//   - add game output channel to selection
// - if user sends game input, pass it to the game
// - if game sends user input, pass it to user
// - if game sends 'other' input, pass it to all users that aren't players
// - if game sends final state, deinitialize game and set status to lobby
func (l *lobby) runLobby() {
	watchdog := time.NewTicker(1 * time.Minute)

	fmt.Println("running lobby", l.name)
	for {
		select {
		case <-watchdog.C:
			if len(l.users) == 0 {
				l.endLobby()
			}
		case userEnded := <-l.userEndConn:
			l.removeUser(userEnded)
			if len(l.users) == 0 {
				l.endLobby()
			}
		case msgFromUser := <-l.webToLobby:
			var msg userMsg
			if err := json.Unmarshal(msgFromUser, &msg); err != nil {
				fmt.Println("user msg error", err)
			}
			if msg.Target == "lobby" {
				l.handleCommand(&msg)
			} else if msg.Target == "game" {
				fmt.Println("lobby passing user message to game")
				if l.g != nil {
					l.lobbyToGame <- msg
				}
				fmt.Println("lobby finished passing user message to game")
			}
		case msgFromGame := <-l.gameToLobby:
			msg, _ := json.Marshal(lobbyMsg{"game", msgFromGame.state})
			if msgFromGame.player == "final" || msgFromGame.player == "" {
				fmt.Println("sending message from game to all players")
				for u := range l.users {
					l.users[u].lobbyToWeb <- msg
				}
				if msgFromGame.player == "final" {
					fmt.Println("game finished, got final output")
					l.handleQuitGame()
				}
			} else {
				fmt.Println("sending message from game to", msgFromGame.player)
				if l.users[msgFromGame.player] != nil {
					l.users[msgFromGame.player].lobbyToWeb <- msg
				}
			}
		}
	}
}
