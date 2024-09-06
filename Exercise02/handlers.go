package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"simplestInterface/db"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	page := 0
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	var db1 db.DataBase
	db1.Name = "places"
	var data ViewData
	data.Places, data.Total, _ = db1.GetPlaces(10, page)
	if page < 0 || page > data.Total/10 {
		w.Write([]byte("Invalid 'page' value: " + pageStr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data.Page = page
	data.Next = fmt.Sprintf("/?page=%d", page+1)
	data.Prev = fmt.Sprintf("/?page=%d", page-1)
	data.Last = fmt.Sprintf("/?page=%d", data.Total/10)
	data.LastInt = data.Total / 10
	tmpl, err := template.ParseFiles("template/template.html")
	if err != nil {
		log.Fatalln("error tempalate.ParseFiles: ", err)
	}
	tmpl.Execute(w, data)
}

func JsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	page := 0
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	var db1 db.DataBase
	db1.Name = "places"
	var data ViewDataJson
	data.Places, data.Total, _ = db1.GetPlaces(10, page)
	data.Name = "places"
	if page > 0 {
		data.Prev = page - 1
	}
	data.Last = data.Total / 10
	if page < 0 || page > data.Last {
		var ed ErrorData
		ed.Error = "Invalid 'page' value: " + pageStr
		w.WriteHeader(http.StatusBadRequest)
		json, _ := json.MarshalIndent(ed, "", " ")
		w.Write(json)
		return
	}
	if page < data.Total {
		data.Next = page + 1
	}
	json, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal("error marchal data: ", err)
	}
	w.Write(json)
}

func RecomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	tokenFromHeader, ok := r.Header["Authorization"]
	if !ok {
		http.Error(w, "Do not Authorized", http.StatusUnauthorized)
		return
	}
	_, tokenString, _ := strings.Cut(tokenFromHeader[0], " ")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret-key"), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Do not Authorized", http.StatusUnauthorized)
		return
	}
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	fmt.Println(latStr," ", lonStr)
	var db1 db.DataBase
	db1.Name = "places"
	db1.Lat, _ = strconv.ParseFloat(latStr, 64)
	db1.Lon, _ = strconv.ParseFloat(lonStr, 64)
	var data ViewDataRec
	data.Name = "Recommendation"
	data.Places, _, _ = db1.GetPlaces(3, 0)
	json, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Fatal("error marchal data: ", err)
	}
	w.Write(json)
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	payload := jwt.MapClaims{
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
		"fuck": "YEAH",
	}
	var token JwtToken
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	strToken, err := t.SignedString([]byte("secret-key"))
	if err != nil {
		log.Fatal("JWT token signing:", err)
	}
	token.Token = strToken
	json, err := json.MarshalIndent(token, "", " ")
	if err != nil {
		log.Fatal("marshal JWT token:", err)
	}
	w.Write(json)
}

func ExitHandler(response http.ResponseWriter, request *http.Request) {
	log.Fatal("/exit handler called")
}
