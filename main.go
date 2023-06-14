package main

import (
	"flag"
	"log"
	"net/http"
)

//default port 8080
var addr = flag.String("addr", ":8080", "http service address")

var hub *Hub //change?

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	//if r.URL.Path != "/" {
	//	http.Error(w, "Not found", http.StatusNotFound)
	//	return
	//}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//http.ServeFile(w, r, "home.html")
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "home.html")
		return
	}
	if r.URL.Path == "/main.css" {
		http.ServeFile(w, r, "main.css")
		return
	}
	if r.URL.Path == "/client.js" {
		http.ServeFile(w, r, "client.js")
		return
	}
}

func main() {
	flag.Parse() //can change to another port
	hub = newHub()
	go hub.run()
	//http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
