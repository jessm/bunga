package main

import (
	"fmt"
	"net/http"
	"text/template"
)

type HomeData struct {
	Name string
}

type LobbyForm struct {
	Name  string `json:"name"`
	Lobby string `json:"lobby"`
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	data := HomeData{Name: "Temp"}
	tmpl.Execute(w, data)
}

func main() {
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", handleHome)

	managerInit()

	fmt.Println("Starting server")
	http.ListenAndServe(":3111", nil)
}
