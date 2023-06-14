package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

var winning = [8][3]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}

var full = false

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	//broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

type Ready struct {
	Tag    string `json:"tag"`
	Symbol int    `json:"symbol"`
}

type Move struct {
	Tag string `json:"tag"`
}

type Update struct {
	Tag string `json:"tag"`
	Index int `json:"index"`
	Symbol int `json:"symbol"`
}

type Moved struct {
	//tag string	//tbd
	Index int	`json:"index"`
	Symbol int	`json:"symbol"`
}

type player struct {
	player *Client
	symbol int
}

func newHub() *Hub {

	return &Hub{
		//broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func rand3() bool {
	return rand.Intn(2) == 0
}

func endGame(board [9]int) bool {
	for _, pose := range winning {
		if board[pose[0]] != 0 && board[pose[1]] != 0 && board[pose[2]] != 0 {
			if board[pose[0]] == board[pose[1]] && board[pose[1]] == board[pose[2]] {
				fmt.Println("the winner is ", board[pose[0]])
				return true
			}
		}
	}
	return false
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if !full {
				if len(h.clients) == 2 {
					full = true
					var players [2]player
					counter := 2
					for c := range h.clients {
						if counter != 0 {
							players[2-counter] = player{player: c}
							counter--
						} else {
							break
						}
					}
					go func(players [2]player) {
						//create new game
						var board   [9]int
						var isTurn int	//index of player

						rand.Seed(time.Now().UnixNano())
						if rand3() {
							players[0].symbol = 1
							players[1].symbol = 2
							isTurn = 1
						} else {
							players[0].symbol = 2
							players[1].symbol = 1
							isTurn = 0
						}
						//later: register game with hub
						for i := 0; i < 2; i++ {
							r := Ready{
								Tag:    "done",
								Symbol: players[i].symbol,
							}
							ready, err := json.Marshal(r)
							if err != nil {
								log.Println(err)
							}
							select {
							case players[i].player.send <- ready:
							default:
								close(players[i].player.send)
								delete(h.clients, players[i].player)
							}
						}
						for i:=0;i<10;i++ {
							if !endGame(board) {
								//send: player, make a move
								m:=Move{
									Tag: "move",
								}
								move, err := json.Marshal(m)
								if err != nil {
									log.Println(err)
								}
								select {
								case _, ok:= <- players[isTurn].player.send:
									if !ok {
										//broadcast game over
										return
									}
								default:
									select {
									case players[isTurn].player.send <- move:
									default:
										close(players[isTurn].player.send)
										delete(h.clients, players[isTurn].player)
									}
								}
								//wait for player
								mvd := Moved{}
								var jsn []byte
								var ok bool
								select {
								case <-players[isTurn].player.receive:
									fmt.Println("There's a cheater among us.")
									return
								default:
									jsn, ok = <-players[isTurn].player.receive
									if !ok {
										fmt.Println("I give up")
										return
									}
								}
								//server duty
								err = json.Unmarshal(jsn, &mvd)
								if err != nil {
									log.Println("Decoding error: ", err)
									return
								}
								board[mvd.Index]=mvd.Symbol
								u := Update{
									Tag:    "update",
									Index: mvd.Index,
									Symbol: mvd.Symbol,
								}
								update, err1 := json.Marshal(u)
								if err1 != nil {
									log.Println(err1)
								}
								for i := 0; i < 2; i++ {
									select {
									case _, ok:= <- players[i].player.send:
										if !ok {
											fmt.Println("Waiting for too long...\nleaving")
											return
										}
									default:
										select {
										case players[i].player.send <- update:
										default:
											close(players[i].player.send)
											delete(h.clients, players[i].player)
										}
									}
								}
								//change turns
								if isTurn==0 {
									isTurn=1
								} else {
									isTurn=0
								}
							} else {
								//todo: send game-over
								break
							}
						}
						//return
					}(players)
				}
			} else {
				err := client.conn.WriteMessage(1, []byte("All spots are taken!")) //to change
				if err != nil {
					log.Println(err)
				}
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				close(client.receive)
			}
			//case message := <-h.broadcast:
			//	for client := range h.clients {
			//		select {
			//		case client.send <- message:
			//		default:
			//			close(client.send)
			//			delete(h.clients, client)
			//		}
			//	}
		}
	}
}