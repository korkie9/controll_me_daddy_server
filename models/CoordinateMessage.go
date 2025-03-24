package models

type CoordinateMessage struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Side string  `json:"side"`
}
