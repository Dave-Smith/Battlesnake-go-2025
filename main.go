package main

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

func main() {
	RunServer()
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
