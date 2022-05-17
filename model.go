package main

// Label is model of the label
type Label struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	LastModifier string `json:"last_modifier"`
	Creator      string `json:"creator"`
}
