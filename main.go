package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
)

type Daten struct {
	Data string `json:"data"`
	Iv   string `json:"iv"`
	Salt string `json:"salt"`
}

var datenmap map[string]string

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		randomIndex, err := randomInt(len(charset))
		if err != nil {
			return "", err
		}
		result[i] = charset[randomIndex]
	}

	return string(result), nil
}

func randomInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func main() {
	datenmap = make(map[string]string)
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)
	http.HandleFunc("/up", handler)
	http.HandleFunc("/data", data)

	err := http.ListenAndServe(":8999", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func data(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fetchkey := r.URL.Query().Get("fetchkey")
	fmt.Printf("fetchkey: %s\n", fetchkey)
	fmt.Fprintln(w, datenmap[fetchkey])
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET")
	case "POST":

		fmt.Println("POST")
		body, _ := io.ReadAll((r.Body))
		defer r.Body.Close()

		var data Daten
		err := json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println(err)
		}
		randomStr, _ := generateRandomString(4)
		datenmap[randomStr] = string(body)
		fmt.Fprintln(w, randomStr)
		fmt.Println(datenmap)

	}
}
