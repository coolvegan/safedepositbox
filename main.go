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

var security map[string]time.Time

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
	security = make(map[string]time.Time)
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

	switch r.Method {
	case "DELETE":
		builder := MongoBuilder{}
		sec := r.Header.Get("X-Sec-Response")
		newtime := time.Now().Add(time.Minute * -2)
		storeTimepoint, ok := security[sec]
		if !ok || newtime.Unix() > storeTimepoint.Unix() {
			fmt.Fprintf(w, "Secret bereits gelöscht.")
			return
		}
		delete(security, sec)
		cfg := Config{Username: os.Getenv("MGUSER"), Password: os.Getenv("MGPASSWORD"), Host: os.Getenv("MGHOST"), Port: os.Getenv("MGPORT"), AuthSource: os.Getenv("MGAUTH"), DatabaseName: os.Getenv("MGDATABASE")}
		builder.Init(&cfg)
		db := builder.Build()
		fetchkey, _ := io.ReadAll((r.Body))
		var data map[string]string
		err := json.Unmarshal(fetchkey, &data)
		if err != nil {
			log.Println(err)
			return
		}
		deleteKey := data["data"]
		defer r.Body.Close()
		err = db.DeleteByKey(deleteKey)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Error")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Secret gelöscht")
	case "GET":
		fetchkey := r.URL.Query().Get("fetchkey")
		w.Header().Set("Content-Type", "application/json")
		builder := MongoBuilder{}
		cfg := Config{Username: os.Getenv("MGUSER"), Password: os.Getenv("MGPASSWORD"), Host: os.Getenv("MGHOST"), Port: os.Getenv("MGPORT"), AuthSource: os.Getenv("MGAUTH"), DatabaseName: os.Getenv("MGDATABASE")}
		builder.Init(&cfg)
		db := builder.Build()

		result, err := db.GetByKey(fetchkey)
		if err != nil {
			log.Println(err)
			fmt.Fprintln(w, "")
		}
		randmstr, _ := generateRandomString(20)
		w.Header().Add("x-sec", randmstr)
		security[randmstr] = time.Now()
		fmt.Fprintln(w, result)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	builder := MongoBuilder{}
	cfg := Config{Username: os.Getenv("MGUSER"), Password: os.Getenv("MGPASSWORD"), Host: os.Getenv("MGHOST"), Port: os.Getenv("MGPORT"), AuthSource: os.Getenv("MGAUTH"), DatabaseName: os.Getenv("MGDATABASE")}
	builder.Init(&cfg)
	db := builder.Build()
	switch r.Method {
	case "POST":
		body, _ := io.ReadAll((r.Body))
		defer r.Body.Close()

		var data SecretStore
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Println(err)
		}
		randomStr, _ := generateRandomString(4)
		datenmap[randomStr] = string(body)
		fmt.Fprintln(w, randomStr)
		data.Created = time.Now()
		data.Code = randomStr

		err = db.Insert(&data)
		if err != nil {
			log.Println(err)
		}

	}
}
