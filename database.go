package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION      = "SecretStore"
	FILTERCOLUMNAME = "code"
)

type DatabaseI interface {
	GetAll() (map[string]string, error)
	GetByKey(key string) (SecretStore, error)
	DeleteByKey(key string) error
	Insert(data *SecretStore) error
}

type Config struct {
	Username     string
	Password     string
	Host         string
	Port         string
	AuthSource   string
	DatabaseName string
}

type MongoBuilder struct {
	cfg Config
}

func (m *MongoBuilder) Init(config *Config) {
	m.cfg = *config
}

func (m *MongoBuilder) Build() *MongoDB {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=%s",
		m.cfg.Username,
		m.cfg.Password,
		m.cfg.Host,
		m.cfg.Port,
		m.cfg.AuthSource,
	)
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Verbindung überprüfen
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database(m.cfg.DatabaseName)
	return &MongoDB{client: client, database: database}
}

func (m *MongoBuilder) Username(username string) {
	m.cfg.Username = username
}

func (m *MongoBuilder) Password(password string) {
	m.cfg.Password = password
}

func (m *MongoBuilder) Host(host string) {
	m.cfg.Host = host
}

func (m *MongoBuilder) Port(port string) {
	m.cfg.Port = port
}

func (m *MongoBuilder) AuthSource(authSource string) {
	m.cfg.AuthSource = authSource
}

func (m *MongoBuilder) DatabaseName(databasename string) {
	m.cfg.DatabaseName = databasename
}

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

func (m *MongoDB) Insert(data *SecretStore) error {
	collection := m.database.Collection(COLLECTION)
	r, err := collection.InsertOne(context.TODO(), data)
	if r.InsertedID == 0 || err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) GetAll() (map[string]string, error) {
	result := make(map[string]string)
	collection := m.database.Collection(COLLECTION)
	cur, err := collection.Find(context.TODO(), bson.M{})
	defer cur.Close(context.TODO())

	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var data SecretStore
		if err := cur.Decode(&data); err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		result[data.Code] = string(b)
	}
	return result, nil
}

func (m *MongoDB) GetByKey(key string) (string, error) {

	collection := m.database.Collection(COLLECTION)
	filter := bson.M{FILTERCOLUMNAME: key}
	set := collection.FindOne(context.TODO(), filter)
	var data SecretStore
	set.Decode(&data)
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	result := string(b)
	if len(result) == 0 {
		return "", fmt.Errorf("Kein Datensatz vorhanden.")
	}
	return result, nil
}

func (m *MongoDB) DeleteByKey(key string) error {

	collection := m.database.Collection(COLLECTION)
	filter := bson.M{FILTERCOLUMNAME: key}
	p, err := collection.DeleteOne(context.TODO(), filter)
	if p.DeletedCount == 0 || err != nil {
		return fmt.Errorf("Datensatz konnte nicht gelöscht werden.")
	}
	return nil
}

func NewMongoDB() *MongoDB {
	//Todo: Change to a more open structure
	//Keine Ahnun ob dieses OOP Konzept in Golang sinnvoll ist.
	//Es gibt keine Destruktoren.
	username := os.Getenv("MGUSER")
	password := os.Getenv("MGPASSWORD")
	host := os.Getenv("MGHOST")
	port := os.Getenv("MGPORT")
	authSource := os.Getenv("MGAUTH") // Die Datenbank, in der die Benutzerinformationen gespeichert sind
	databaseName := os.Getenv("MGDATABASE")
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=%s",
		username,
		password,
		host,
		port,
		authSource,
	)
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Verbindung überprüfen
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database(databaseName)
	return &MongoDB{client: client, database: database}
}
