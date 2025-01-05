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

	// Calculate opponent predictions with more advanced analysis
	predictions := newOpponentPredictor(gameState).getPredictions()

	// Calculate scores for each possible move
	for _, direction := range possibleMoves {
		nextPos := getNextPosition(myHead, direction)

		// Skip invalid moves
		if !isValidMove(nextPos, gameState) {
			continue
		}

		// Check for potential head-to-head collisions using advanced prediction
		collisionRisk := calculateCollisionRisk(nextPos, predictions, myLength, gameState)
		if collisionRisk >= 0.8 { // High risk threshold
			continue
		}

		score := evaluateMove(nextPos, gameState, myHealth, myLength)

		// Adjust score based on collision risk
		score *= (1.0 - collisionRisk)

		validMoves = append(validMoves, Move{Direction: direction, Score: score})
	}

	// If no valid moves, try to accept moves with higher risk (better than guaranteed death)
	if len(validMoves) == 0 {
		for _, direction := range possibleMoves {
			nextPos := getNextPosition(myHead, direction)
			if !isValidMove(nextPos, gameState) {
				continue
			}

			collisionRisk := calculateCollisionRisk(nextPos, predictions, myLength, gameState)
			score := evaluateMove(nextPos, gameState, myHealth, myLength)

			// Apply risk-based penalty
			score *= (1.0 - collisionRisk)
			score -= 200 // Additional penalty for high-risk moves

			validMoves = append(validMoves, Move{Direction: direction, Score: score})
		}
	}

	if len(validMoves) == 0 {
		return "up"
	}

	bestMove := validMoves[0]
	for _, move := range validMoves {
		if move.Score > bestMove.Score {
			bestMove = move
		}
	}

	return bestMove.Direction
}

// OpponentPredictor provides advanced opponent movement prediction
type OpponentPredictor struct {
	gameState GameState
	cache     map[string]PredictionData
}

type PredictionData struct {
	LikelyMoves     []Coordinate
	MoveProbability map[Coordinate]float64
	Intent          MovementIntent
}

type MovementIntent struct {
	SeekingFood    bool
	AggressiveMode bool
	Trapped        bool
}

func newOpponentPredictor(state GameState) *OpponentPredictor {
	return &OpponentPredictor{
		gameState: state,
		cache:     make(map[string]PredictionData),
	}
}

func (op *OpponentPredictor) getPredictions() map[string]PredictionData {
	for _, snake := range op.gameState.Board.Snakes {
		if snake.ID == op.gameState.You.ID {
			continue
		}

		if _, exists := op.cache[snake.ID]; !exists {
			op.cache[snake.ID] = op.analyzeSnake(snake)
		}
	}
	return op.cache
}

func (op *OpponentPredictor) analyzeSnake(snake Snake) PredictionData {
	data := PredictionData{
		MoveProbability: make(map[Coordinate]float64),
	}

	// Determine snake's intent
	intent := op.determineIntent(snake)
	data.Intent = intent

	// Get all possible moves
	directions := []string{"up", "down", "left", "right"}
	totalWeight := 0.0

	for _, dir := range directions {
		nextPos := getNextPosition(snake.Head, dir)
		if !isValidMove(nextPos, op.gameState) {
			continue
		}

		weight := op.calculateMoveWeight(nextPos, snake, intent)
		if weight > 0 {
			data.LikelyMoves = append(data.LikelyMoves, nextPos)
			data.MoveProbability[nextPos] = weight
			totalWeight += weight
		}
	}

	// Normalize probabilities
	if totalWeight > 0 {
		for pos := range data.MoveProbability {
			data.MoveProbability[pos] /= totalWeight
		}
	}

	return data
}

func (op *OpponentPredictor) determineIntent(snake Snake) MovementIntent {
	intent := MovementIntent{}

	// Check if snake is seeking food
	if snake.Health < 30 {
		intent.SeekingFood = true
	}

	// Check if snake is in aggressive mode
	for _, otherSnake := range op.gameState.Board.Snakes {
		if otherSnake.ID == snake.ID {
			continue
		}
		headDist := manhattanDistance(snake.Head, otherSnake.Head)
		if headDist <= 2 && snake.Length > otherSnake.Length {
			intent.AggressiveMode = true
			break
		}
	}

	// Check if snake is trapped
	availableMoves := 0
	for _, dir := range []string{"up", "down", "left", "right"} {
		nextPos := getNextPosition(snake.Head, dir)
		if isValidMove(nextPos, op.gameState) {
			availableMoves++
		}
	}
	intent.Trapped = availableMoves <= 2

	return intent
}

func (op *OpponentPredictor) calculateMoveWeight(pos Coordinate, snake Snake, intent MovementIntent) float64 {
	weight := 1.0

	// Adjust weight based on food proximity if snake is seeking food
	if intent.SeekingFood {
		minFoodDist := math.MaxFloat64
		for _, food := range op.gameState.Board.Food {
			dist := float64(manhattanDistance(pos, food))
			if dist < minFoodDist {
				minFoodDist = dist
			}
		}
		weight *= (10.0 / (minFoodDist + 1.0))
	}

	// Adjust weight based on aggressive behavior
	if intent.AggressiveMode {
		for _, otherSnake := range op.gameState.Board.Snakes {
			if otherSnake.ID == snake.ID || otherSnake.ID == op.gameState.You.ID {
				continue
			}
			if snake.Length > otherSnake.Length {
				headDist := float64(manhattanDistance(pos, otherSnake.Head))
				weight *= (5.0 / (headDist + 1.0))
			}
		}
	}

	// Reduce weight for moves that could trap the snake
	if intent.Trapped {
		availableNextMoves := 0
		for _, dir := range []string{"up", "down", "left", "right"} {
			nextPos := getNextPosition(pos, dir)
			if isValidMove(nextPos, op.gameState) {
				availableNextMoves++
			}
		}
		weight *= float64(availableNextMoves) / 4.0
	}

	return weight
}

func calculateCollisionRisk(nextPos Coordinate, predictions map[string]PredictionData, myLength int, state GameState) float64 {
	maxRisk := 0.0

	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}

		prediction := predictions[snake.ID]

		// Calculate collision risk based on move probabilities
		for possiblePos, probability := range prediction.MoveProbability {
			if nextPos.X == possiblePos.X && nextPos.Y == possiblePos.Y {
				risk := probability

				// Adjust risk based on snake lengths
				if snake.Length >= myLength {
					risk *= 1.5 // Increase risk for longer or equal length snakes
				} else {
					risk *= 0.5 // Decrease risk for shorter snakes
				}

				// Adjust risk based on snake's intent
				if prediction.Intent.AggressiveMode {
					risk *= 1.3 // Increase risk if snake is in aggressive mode
				}
				if prediction.Intent.Trapped {
					risk *= 0.7 // Decrease risk if snake is trapped
				}

				if risk > maxRisk {
					maxRisk = risk
				}
			}
		}
	}

	return maxRisk
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
