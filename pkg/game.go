package main

type game interface {
	runGame()
	broadcastState()
}

type gameMsg struct {
	player string
	state  interface{}
}
