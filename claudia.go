package main

import (
	"fmt"
	"math"
	"sort"
)

type Move struct {
	Direction string
	Score     float64
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

// evaluateMove scores a potential move based on various factors with enhanced food strategy
func evaluateMove(pos Coordinate, state GameState, myHealth int, myLength int) float64 {
	// Base score
	score := 100.0

	// Initialize safety multiplier
	safetyMultiplier := 1.0

	// -------- CRITICAL SAFETY CHECKS (Massive Penalties) --------

	// Check for immediate head-to-head possibilities with larger or equal snakes
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}

		headDist := manhattanDistance(pos, snake.Head)
		if headDist == 1 && snake.Length >= myLength {
			return -1000.0 // Extremely negative score to avoid certain death
		}
	}

	// Check if the move leads to a potential trap
	if isTrappedPosition(pos, state, 3) {
		return -800.0
	}

	// -------- HAZARD AVOIDANCE --------

	for _, hazard := range state.Board.Hazards {
		if pos.X == hazard.X && pos.Y == hazard.Y {
			safetyMultiplier *= 0.5
		}
	}

	// -------- FOOD EVALUATION WITH LENGTH STRATEGY --------

	// Calculate optimal length based on other snakes
	optimalLength := calculateOptimalLength(state)

	// Initialize food score component
	foodScore := 0.0

	// Find closest food
	closestFoodDist := math.MaxFloat64
	var closestFood *Coordinate

	for _, food := range state.Board.Food {
		dist := float64(manhattanDistance(pos, food))
		if dist < closestFoodDist {
			closestFoodDist = dist
			foodClone := food
			closestFood = &foodClone
		}
	}

	if closestFood != nil {
		// Check if the path to food is safe
		isFoodSafe := true
		if closestFoodDist == 0 { // We're about to eat food
			for _, snake := range state.Board.Snakes {
				if snake.ID == state.You.ID {
					continue
				}
				snakeToFoodDist := manhattanDistance(snake.Head, *closestFood)
				if snakeToFoodDist == 1 && snake.Length >= myLength {
					isFoodSafe = false
					break
				}
			}
		}

		// Determine if we should seek food based on length strategy
		shouldSeekFood := false
		urgentFood := false

		if myHealth < 25 {
			// Emergency food seeking
			urgentFood = true
			shouldSeekFood = true
		} else if myHealth < 50 {
			// Check if we're below optimal length
			if myLength < optimalLength {
				shouldSeekFood = true
			}
		} else {
			// Only seek food if we're significantly below optimal length
			if myLength < optimalLength-2 {
				shouldSeekFood = true
			}
		}

		// Calculate food score based on strategy
		if shouldSeekFood && isFoodSafe {
			if urgentFood {
				foodScore = 300.0 / (closestFoodDist + 1)
			} else {
				foodScore = 150.0 / (closestFoodDist + 1)
			}
		} else if myLength > optimalLength {
			// Slightly avoid food when we're already longer than optimal
			foodScore = -20.0 / (closestFoodDist + 1)
		}
	}

	// -------- SPACE EVALUATION --------

	spaceScore := evaluateAvailableSpace(pos, state, 5)

	// Weight space more heavily when we're at or above optimal length
	if myLength >= optimalLength {
		score += spaceScore * 75 // Increased weight on space when we're long enough
	} else {
		score += spaceScore * 50
	}

	// -------- TAIL CHASING BEHAVIOR --------

	// Encourage tail chasing when we're at or above optimal length and not hungry
	if myLength >= optimalLength && myHealth > 50 {
		tailDist := manhattanDistance(pos, state.You.Body[len(state.You.Body)-1])
		if tailDist <= 2 {
			score += 100.0 / (float64(tailDist) + 1)
		}
	}

	// -------- AGGRESSIVE/DEFENSIVE POSITIONING --------

	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}

		headDist := manhattanDistance(pos, snake.Head)

		if myLength > snake.Length+1 {
			// Aggressive positioning towards smaller snakes
			if headDist == 2 {
				score += 50.0
			}
		} else {
			// Defensive positioning against larger snakes
			if headDist <= 2 {
				safetyMultiplier *= 0.7
			}
		}
	}

	// -------- FINAL SCORE CALCULATION --------

	score = (score + foodScore) * safetyMultiplier

	return score
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

// calculateOptimalLength determines the ideal length based on other snakes
func calculateOptimalLength(state GameState) int {
	var snakeLengths []int
	var maxLength int
	var avgLength float64
	totalSnakes := 0

	// Collect snake lengths and find max
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		snakeLengths = append(snakeLengths, snake.Length)
		if snake.Length > maxLength {
			maxLength = snake.Length
		}
		avgLength += float64(snake.Length)
		totalSnakes++
	}

	if totalSnakes == 0 {
		return 3 // Minimum length if no other snakes
	}

	avgLength /= float64(totalSnakes)

	// Calculate optimal length based on game state
	optimalLength := int(avgLength) + 1

	// Ensure optimal length is at least slightly longer than shortest snake
	minDesiredLength := 3
	if len(snakeLengths) > 0 {
		sort.Ints(snakeLengths)
		minDesiredLength = snakeLengths[0] + 1
	}

	// Don't try to get too long - being too long makes it harder to maneuver
	maxDesiredLength := maxLength
	if maxDesiredLength > 8 {
		maxDesiredLength = 8 // Cap maximum desired length
	}

	// Adjust optimal length to stay within bounds
	if optimalLength < minDesiredLength {
		optimalLength = minDesiredLength
	}
	if optimalLength > maxDesiredLength {
		optimalLength = maxDesiredLength
	}

	return optimalLength
}

// isValidMove checks if the move is valid (within bounds and not colliding)
// Now considers snake tails that will move next turn
func isValidMove(pos Coordinate, state GameState) bool {
	// Check board boundaries
	if pos.X < 0 || pos.X >= state.Board.Width || pos.Y < 0 || pos.Y >= state.Board.Height {
		return false
	}

	// Check collision with all snake bodies, excluding tails that will move
	for _, snake := range state.Board.Snakes {
		for i, segment := range snake.Body {
			// Skip the tail piece unless it's going to grow
			// A snake grows when it eats food, which happens when its head is on food
			isSnakeGrowing := false
			for _, food := range state.Board.Food {
				if snake.Head.X == food.X && snake.Head.Y == food.Y {
					isSnakeGrowing = true
					break
				}
			}

			// If this is the tail and the snake isn't growing, the space will be free next turn
			if i == len(snake.Body)-1 && !isSnakeGrowing {
				continue
			}

			if pos.X == segment.X && pos.Y == segment.Y {
				return false
			}
		}
	}

	return true
}

// evaluateAvailableSpace calculates how much free space is available from a position
func evaluateAvailableSpace(pos Coordinate, state GameState, depth int) float64 {
	visited := make(map[string]bool)

	// Create a modified state for flood fill that considers tail positions as free
	modifiedState := createStateWithMovedTails(state)

	return float64(floodFill(pos, modifiedState, depth, visited))
}

// createStateWithMovedTails creates a copy of the game state where non-growing tail positions are marked as free
func createStateWithMovedTails(state GameState) GameState {
	modifiedState := state
	modifiedState.Board.Snakes = make([]Snake, len(state.Board.Snakes))

	for i, snake := range state.Board.Snakes {
		// Check if snake will grow
		isGrowing := false
		for _, food := range state.Board.Food {
			if snake.Head.X == food.X && snake.Head.Y == food.Y {
				isGrowing = true
				break
			}
		}

		// Copy snake but remove tail if not growing
		newSnake := snake
		if !isGrowing && len(snake.Body) > 0 {
			newSnake.Body = make([]Coordinate, len(snake.Body)-1)
			copy(newSnake.Body, snake.Body[:len(snake.Body)-1])
		} else {
			newSnake.Body = make([]Coordinate, len(snake.Body))
			copy(newSnake.Body, snake.Body)
		}
		modifiedState.Board.Snakes[i] = newSnake
	}

	return modifiedState
}

// isTrappedPosition checks if a position might lead to being trapped
func isTrappedPosition(pos Coordinate, state GameState, depth int) bool {
	visited := make(map[string]bool)

	// Use modified state that considers tail positions
	modifiedState := createStateWithMovedTails(state)
	availableSpace := floodFill(pos, modifiedState, depth, visited)

	// Consider it trapped if there's very limited space
	// Increased minimum space requirement to account for tail positions
	return availableSpace < (depth*2 + 1)
}

// floodFill remains the same but uses the modified state
func floodFill(pos Coordinate, state GameState, depth int, visited map[string]bool) int {
	if depth == 0 {
		return 0
	}

	key := fmt.Sprintf("%d,%d", pos.X, pos.Y)
	if visited[key] {
		return 0
	}

	if !isValidMove(pos, state) {
		return 0
	}

	visited[key] = true
	count := 1

	// Check all directions
	directions := []string{"up", "down", "left", "right"}
	for _, dir := range directions {
		nextPos := getNextPosition(pos, dir)
		count += floodFill(nextPos, state, depth-1, visited)
	}

	return count
}
