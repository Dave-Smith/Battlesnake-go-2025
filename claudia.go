package main

import (
	"fmt"
	"math"
	"sort"
)

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
