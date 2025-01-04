package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type SnakeMoverFunc func(state GameState) BattlesnakeMoveResponse
type SnakeStartFunc func(state GameState)
type SnakeInfoFunc func() BattlesnakeInfoResponse
type SnakeEndFunc func(state GameState)

func HandleStart(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		log.Printf("ERROR: Failed to decode start json, %s", err)
		return
	}

	// Nothing to respond with here
}

// HandleMove processes the move request and returns the next move
func HandleMove(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var gameState GameState
	if err := json.NewDecoder(r.Body).Decode(&gameState); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate the next move
	nextMove := calculateNextMove(gameState)

	w.Header().Set("Server", ServerID)
	// Create the response
	response := MoveResponse{
		Move:  nextMove,
		Shout: "Going " + nextMove + "!",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Send the response
	json.NewEncoder(w).Encode(response)
}

func HandleEnd(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		log.Printf("ERROR: Failed to decode end json, %s", err)
		return
	}

	// Nothing to respond with here
}

// Middleware

const ServerID = "battlesnake/dave-smith/claudia"

func withServerID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", ServerID)
		if next != nil {
			next(w, r)
		}
	}
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "Dave-Smith",
		Color:      "#7ABF36",
		Head:       "all-seeing",
		Tail:       "do-sammy",
		Version:    "0.0.1",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("ERROR: Failed to encode info response, %s", err)
	}
}

func SnakeHandlerMove(mover SnakeMoverFunc, serverId string, next http.HandlerFunc) http.HandlerFunc {
	log.Printf("Move")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", serverId)
		if next != nil {
			next(w, r)
		}

		state, err := unmarshalState(r)
		if err != nil {
			log.Printf("ERROR: Failed to decode move json, %s", err)
			return
		}
		log.Printf("[%s] Head position: (%d,%d), Body: %v, Health: %d, Length: %d", state.You.Name, state.You.Head.X, state.You.Head.Y, state.You.Body, state.You.Health, state.You.Length)

		response := mover(state)

		log.Printf("[%s] Moving %s", state.You.Name, response.Move)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("ERROR: Failed to encode move response, %s", err)
			return
		}
	}
}

func SnakeHandlerStart(starter SnakeStartFunc, serverId string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", serverId)
		if next != nil {
			next(w, r)
		}
		state, err := unmarshalState(r)
		log.Printf("[%s] Starting new game", state.You.Name)
		if err != nil {
			log.Printf("ERROR: Failed to decode move json, %s", err)
			return
		}
		log.Printf("[%s] Head position: (%d,%d), Body: %v, Health: %d, Length: %d", state.You.Name, state.You.Head.X, state.You.Head.Y, state.You.Body, state.You.Health, state.You.Length)

		starter(state)

		return
	}
}

func SnakeHandlerInfo(snakeInfo SnakeInfoFunc, serverId string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", serverId)

		response := snakeInfo()

		if next != nil {
			next(w, r)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("ERROR: Failed to encode info response, %s", err)
		}
	}
}

func SnakeHandlerEnd(gameEnd SnakeEndFunc, serverId string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := unmarshalState(r)
		if err != nil {
			log.Printf("ERROR: Failed to decode move json, %s", err)
			return
		}
		gameEnd(state)
	}
}

func unmarshalState(r *http.Request) (GameState, error) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		return GameState{}, err
	}
	return state, nil
}

// Start Battlesnake Server
func RunServer() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	http.HandleFunc("/claudia/", withServerID(HandleIndex))
	http.HandleFunc("/claudia/start", withServerID(HandleStart))
	http.HandleFunc("/claudia/move", withServerID(HandleMove))
	http.HandleFunc("/claudia/end", withServerID(HandleEnd))

	log.Printf("Running Battlesnake at http://0.0.0.0:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
