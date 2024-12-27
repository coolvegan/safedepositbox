package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
)

type SecretStore struct {
	Data    string    `json:"data" bson:"data"`
	Iv      string    `json:"iv" bson:"iv"`
	Salt    string    `json:"salt" bson:"salt"`
	Created time.Time `bson:"timestamp"`
	Code    string    `bson:"code"`
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
	builder := MongoBuilder{}
	cfg := Config{Username: os.Getenv("MGUSER"), Password: os.Getenv("MGPASSWORD"), Host: os.Getenv("MGHOST"), Port: os.Getenv("MGPORT"), AuthSource: os.Getenv("MGAUTH"), DatabaseName: os.Getenv("MGDATABASE")}
	builder.Init(&cfg)
	db := builder.Build()
	result, err := db.GetAll()
	if err != nil {
		log.Fatal("Can't Access Database.")
	}
	datenmap = result
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)
	http.HandleFunc("/up", handler)
	http.HandleFunc("/data", data)

	err = http.ListenAndServe("127.0.0.1:8999", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func data(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fetchkey := r.URL.Query().Get("fetchkey")
	builder := MongoBuilder{}
	cfg := Config{Username: os.Getenv("MGUSER"), Password: os.Getenv("MGPASSWORD"), Host: os.Getenv("MGHOST"), Port: os.Getenv("MGPORT"), AuthSource: os.Getenv("MGAUTH"), DatabaseName: os.Getenv("MGDATABASE")}
	builder.Init(&cfg)
	db := builder.Build()

	result, err := db.GetByKey(fetchkey)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(w, "")
	}
	fmt.Fprintln(w, result)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		body, _ := io.ReadAll((r.Body))
		defer r.Body.Close()
		fmt.Println("error" + string(body))

		var data SecretStore
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Println(err)
		}
		randomStr, _ := generateRandomString(4)
		datenmap[randomStr] = string(body)
		fmt.Fprintln(w, randomStr)
		fmt.Println(datenmap)
		data.Created = time.Now()
		data.Code = randomStr

		builder := MongoBuilder{}
		cfg := Config{Username: os.Getenv("MGUSER"), Password: os.Getenv("MGPASSWORD"), Host: os.Getenv("MGHOST"), Port: os.Getenv("MGPORT"), AuthSource: os.Getenv("MGAUTH"), DatabaseName: os.Getenv("MGDATABASE")}
		builder.Init(&cfg)
		db := builder.Build()
		err = db.Insert(&data)
		if err != nil {
			log.Println(err)
		}

	}
}
