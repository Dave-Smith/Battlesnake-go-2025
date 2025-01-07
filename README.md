# Battlesnake Project

This project implements a Battlesnake AI using Go. The AI is designed to control a snake in the Battlesnake game, making strategic decisions to survive and compete against other snakes. Below is a summary of the key features and enhancements implemented in this project.

## Features

### Movement Logic
- **Basic Movement**: The snake can move in four directions: up, down, left, and right.
- **Boundary Avoidance**: Ensures the snake does not move off the board.
- **Collision Avoidance**: Prevents the snake from running into its own body or other snakes.

### Food Seeking
- **Health-Based Food Seeking**: The snake seeks food only when its health is low (< 25). When health is medium (< 50), it seeks food if it is safe. When health is high, it avoids food unless it is very safe and nearby.
- **Optimal Length Strategy**: The snake aims to maintain an optimal length, slightly longer than the shortest opponent but not excessively long to maintain maneuverability.

### Opponent Interaction
- **Aggressive Behavior**: The snake attacks nearby opponents when they have a shorter length.
- **Defensive Positioning**: The snake avoids head-to-head collisions with larger or equal-length snakes.

### Advanced Prediction
- **Opponent Prediction**: Uses an `OpponentPredictor` to analyze opponent snakes' likely moves based on their current state and behavior patterns.
- **Collision Risk Calculation**: Calculates collision risk based on opponent move probabilities, snake lengths, and behavioral patterns.

### Space Evaluation
- **Flood Fill Algorithm**: Evaluates available space from a position to avoid getting trapped or cornered.
- **Tail Position Handling**: Considers tail positions of other snakes that will move next turn, preventing the snake from mistaking these spaces for dead ends.

### Safety Enhancements
- **Critical Safety Checks**: Immediate disqualification of moves that lead to certain death, such as head-to-head collisions with larger snakes or moves into trapped positions.
- **Hazard Avoidance**: Applies penalties for moving into hazardous areas on the board.

### HTTP Handlers
- **Move Response**: Handles HTTP requests to determine the next move and returns a JSON response with the move and a shout message.
- **Start, End, and Info Handlers**: Handles game start, end, and info requests.

## Usage

To use this code in your Battlesnake project:
1. Add the provided functions and handlers to your project.
2. Set up the HTTP server to handle requests at the appropriate endpoints.
3. Parse incoming requests into the `GameState` struct and call `calculateNextMove` to determine the next move.
4. Return the move in the required JSON format.

Example setup for the HTTP server:
```go
func main() {
    http.HandleFunc("/move", HandleMove)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

This project aims to create a competitive and strategic Battlesnake AI that can adapt to various game scenarios and opponents.