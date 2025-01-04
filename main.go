package main

import (
	"math"
)

// MoveResponse represents the response structure required by Battlesnake API
type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout"`
}

// Rest of the code remains the same...
type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Move struct {
	Direction string
	Score     float64
}

func main() {
	RunServer()
}

// calculateNextMove determines the best move for the snake
func calculateNextMove(gameState GameState) string {
	possibleMoves := []string{"up", "down", "left", "right"}
	var validMoves []Move

	myHead := gameState.You.Head
	myHealth := gameState.You.Health
	myLength := gameState.You.Length

	// Calculate scores for each possible move
	for _, direction := range possibleMoves {
		nextPos := getNextPosition(myHead, direction)

		// Skip invalid moves
		if !isValidMove(nextPos, gameState) {
			continue
		}

		score := evaluateMove(nextPos, gameState, myHealth, myLength)
		validMoves = append(validMoves, Move{Direction: direction, Score: score})
	}

	// If no valid moves, return any direction (snake will die anyway)
	if len(validMoves) == 0 {
		return "up"
	}

	// Return the move with the highest score
	bestMove := validMoves[0]
	for _, move := range validMoves {
		if move.Score > bestMove.Score {
			bestMove = move
		}
	}

	return bestMove.Direction
}

// getNextPosition calculates the next position based on current position and direction
func getNextPosition(current Coordinate, direction string) Coordinate {
	switch direction {
	case "up":
		return Coordinate{X: current.X, Y: current.Y + 1}
	case "down":
		return Coordinate{X: current.X, Y: current.Y - 1}
	case "left":
		return Coordinate{X: current.X - 1, Y: current.Y}
	case "right":
		return Coordinate{X: current.X + 1, Y: current.Y}
	}
	return current
}

// isValidMove checks if the move is valid (within bounds and not colliding)
func isValidMove(pos Coordinate, state GameState) bool {
	// Check board boundaries
	if pos.X < 0 || pos.X >= state.Board.Width || pos.Y < 0 || pos.Y >= state.Board.Height {
		return false
	}

	// Check collision with all snake bodies
	for _, snake := range state.Board.Snakes {
		for _, segment := range snake.Body {
			if pos.X == segment.X && pos.Y == segment.Y {
				return false
			}
		}
	}

	return true
}

// evaluateMove scores a potential move based on various factors
func evaluateMove(pos Coordinate, state GameState, myHealth int, myLength int) float64 {
	score := 100.0

	// Avoid hazards
	for _, hazard := range state.Board.Hazards {
		if pos.X == hazard.X && pos.Y == hazard.Y {
			score -= 50
		}
	}

	// Food evaluation
	for _, food := range state.Board.Food {
		dist := manhattanDistance(pos, food)
		if myHealth < 25 {
			// Actively seek food when health is low
			score += 50.0 / float64(dist)
		} else {
			// Slightly avoid food when health is good
			score -= 10.0 / float64(dist)
		}
	}

	// Evaluate opponent positions
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}

		// Distance to opponent's head
		headDist := manhattanDistance(pos, snake.Head)

		if headDist == 1 {
			if myLength > snake.Length {
				// Aggressive behavior when we're longer
				score += 100
			} else {
				// Avoid head-to-head with longer or equal length snakes
				score -= 100
			}
		}
	}

	// Evaluate dead ends
	deadEndPenalty := evaluateDeadEnd(pos, state, 4)
	score -= float64(deadEndPenalty)

	return score
}

// manhattanDistance calculates the Manhattan distance between two points
func manhattanDistance(a, b Coordinate) int {
	return int(math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y)))
}

// evaluateDeadEnd checks if a position leads to a dead end
func evaluateDeadEnd(pos Coordinate, state GameState, depth int) float64 {
	if depth == 0 {
		return 0
	}

	availableMoves := 0
	totalScore := 0.0

	// Check all directions
	directions := []string{"up", "down", "left", "right"}
	for _, dir := range directions {
		nextPos := getNextPosition(pos, dir)
		if isValidMove(nextPos, state) {
			availableMoves++
			totalScore += evaluateDeadEnd(nextPos, state, depth-1)
		}
	}

	if availableMoves == 0 {
		return 100.0 // Dead end penalty
	}

	return totalScore / float64(availableMoves)
}

// Required GameState structures
type GameState struct {
	Game  Game  `json:"game"`
	Turn  int   `json:"turn"`
	Board Board `json:"board"`
	You   Snake `json:"you"`
}

type Game struct {
	ID      string `json:"id"`
	Ruleset struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Settings struct {
			FoodSpawnChance     int `json:"foodSpawnChance"`
			MinimumFood         int `json:"minimumFood"`
			HazardDamagePerTurn int `json:"hazardDamagePerTurn"`
		} `json:"settings"`
	} `json:"ruleset"`
}

type Board struct {
	Height  int          `json:"height"`
	Width   int          `json:"width"`
	Food    []Coordinate `json:"food"`
	Hazards []Coordinate `json:"hazards"`
	Snakes  []Snake      `json:"snakes"`
}

type Snake struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Health int          `json:"health"`
	Body   []Coordinate `json:"body"`
	Head   Coordinate   `json:"head"`
	Length int          `json:"length"`
}

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
	Version    string `json:"version"`
}

type BattlesnakeMoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout"`
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Battlesnake struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Health         int            `json:"health"`
	Body           []Coord        `json:"body"`
	Head           Coord          `json:"head"`
	Length         int            `json:"length"`
	Latency        string         `json:"latency"`
	Shout          string         `json:"shout"`
	Customizations Customizations `json:"customizations"`
}

type Customizations struct {
	Color string `json:"color"`
	Head  string `json:"head"`
	Tail  string `json:"tail"`
}
