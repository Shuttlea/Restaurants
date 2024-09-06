package main

import (
	"simplestInterface/db"
)

type ViewData struct {
	Total   int
	Places  []db.Place
	Prev    string
	Next    string
	Last    string
	LastInt int
	Page    int
}

type ViewDataJson struct {
	Name   string     `json:"name"`
	Total  int        `json:"total"`
	Places []db.Place `json:"places"`
	Prev   int        `json:"prev_page"`
	Next   int        `json:"next_page"`
	Last   int        `json:"last_page"`
}

type ViewDataRec struct {
	Name   string     `json:"name"`
	Places []db.Place `json:"places"`
}

type ErrorData struct {
	Error string `json:"error"`
}

type Store interface {
	GetPlaces(limit int, offset int) ([]db.Place, int, error)
}

type JwtToken struct {
	Token string `json:"token"`
}
