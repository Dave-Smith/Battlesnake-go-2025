package main

import (
	"fmt"
	"math"
)

// evaluateMove scores a potential move based on various factors with enhanced safety
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
    if isTrappedPosition(pos, state, 3) { // Check 3 moves ahead
        return -800.0
    }

    // -------- HAZARD AVOIDANCE --------

    // Avoid hazards unless absolutely necessary
    for _, hazard := range state.Board.Hazards {
        if pos.X == hazard.X && pos.Y == hazard.Y {
            safetyMultiplier *= 0.5 // Significant penalty for hazards
        }
    }

    // -------- FOOD EVALUATION --------

    // Initialize food score component
    foodScore := 0.0

    // Find closest food
    closestFoodDist := math.MaxFloat64
    var closestFood *Coordinate

    for _, food := range state.Board.Food {
        dist := float64(manhattanDistance(pos, food))
        if dist < closestFoodDist {
            closestFoodDist = dist
            foodClone := food // Create a copy to avoid pointer issues
            closestFood = &foodClone
        }
    }

    // Only consider food if we really need it or it's safely accessible
    if closestFood != nil {
        // Check if the path to food is safe
        isFoodSafe := true
        if closestFoodDist == 0 { // We're about to eat food
            // Check if any larger/equal snake can also reach this food next turn
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

        // Adjust food score based on health and safety
        if myHealth < 25 {
            // Desperate for food
            foodScore = 200.0 / (closestFoodDist + 1)
            if !isFoodSafe {
                foodScore *= 0.5 // Still risky but might be necessary
            }
        } else if myHealth < 50 {
            // Want food but not desperate
            if isFoodSafe {
                foodScore = 100.0 / (closestFoodDist + 1)
            }
        } else {
            // Don't really need food
            if isFoodSafe && closestFoodDist <= 2 {
                foodScore = 50.0 / (closestFoodDist + 1) // Might as well grab it if it's safe and close
            }
        }
    }

    // -------- SPACE EVALUATION --------

    // Calculate available space in each direction
    spaceScore := evaluateAvailableSpace(pos, state, 5) // Look 5 moves ahead
    score += spaceScore * 50 // Weight space heavily in decision making

    // -------- AGGRESSIVE/DEFENSIVE POSITIONING --------

    for _, snake := range state.Board.Snakes {
        if snake.ID == state.You.ID {
            continue
        }

        headDist := manhattanDistance(pos, snake.Head)

        if myLength > snake.Length + 1 {
            // Aggressive positioning towards smaller snakes
            if headDist == 2 {
                score += 50.0 // Encourage cutting off smaller snakes
            }
        } else {
            // Defensive positioning against larger snakes
            if headDist <= 2 {
                safetyMultiplier *= 0.7 // Reduce score when near larger snakes
            }
        }
    }

    // -------- FINAL SCORE CALCULATION --------

    // Combine all factors with safety multiplier
    score = (score + foodScore) * safetyMultiplier

    return score
}

// evaluateAvailableSpace calculates how much free space is available from a position
func evaluateAvailableSpace(pos Coordinate, state GameState, depth int) float64 {
    visited := make(map[string]bool)
    return float64(floodFill(pos, state, depth, visited))
}

// floodFill counts available spaces using flood fill algorithm
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

// isTrappedPosition checks if a position might lead to being trapped
func isTrappedPosition(pos Coordinate, state GameState, depth int) bool {
    visited := make(map[string]bool)
    availableSpace := floodFill(pos, state, depth, visited)

    // Consider it trapped if there's very limited space
    return availableSpace < (depth * 2)
}
