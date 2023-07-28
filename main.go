package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	uri        = "mongodb://root:example@mongo:27017"
	database   = "Sensor"
	collection = "Sensor"
)

func main() {
	if run() != 0 {
		return
	}
}

func run() int {
	const (
		success = iota
		failure
	)

	opt := options.Client().ApplyURI(uri).SetConnectTimeout(5 * time.Second).SetMaxConnecting(5)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fmt.Printf("connect to mongo: %s\n", uri)

	client, err := mongo.Connect(ctx, opt)
	defer func() {
		err = client.Disconnect(ctx)
		if err != nil {
			fmt.Printf("failed to disconnect mongo: %v\n", err)
		}
	}()

	if err != nil {
		fmt.Printf("failed to connect to mongo: %v\n", err)
		return failure
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Printf("failed to ping mongo: %v\n", err)
		return failure
	}
	fmt.Println("Connect Success.")

	go func() {
		// insert and get length by 3 seconds interval
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sensor, err := getSensor()
				if err != nil {
					fmt.Printf("failed to get sensor data: %v\n", err)
				}
				err = insertSensorData(ctx, client, *sensor)
				if err != nil {
					fmt.Printf("failed to insert sensor data: %v\n", err)
				}
			}
		}
	}()

	go func() {
		// get length by 5 seconds interval
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				count, err := getCount(ctx, client)
				if err != nil {
					fmt.Printf("failed to getCount: %v\n", err)
				}
				fmt.Printf("sensor data length: %d\n", count)
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		// get length by 5 seconds interval
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sensors, err := findSensor(ctx, client)
				if err != nil {
					fmt.Printf("failed to findSensor: %v\n", err)
				}
				fmt.Printf("sensors: %v\n", sensors)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()

	fmt.Println("finished")
	if err != nil {
		fmt.Printf("failed to wait: %v", err)
		return failure
	}

	return success
}

type Sensor struct {
	ID       string `bson:"_id"`
	Humidity int    `bson:"humidity"`
	Temp     int    `bson:"temp"`
	HasSent  bool   `bson:"hasSent"`
}

func NewSensor() *Sensor {
	return &Sensor{
		ID:       uuid.New().String(),
		Humidity: rand.Intn(100),
		Temp:     rand.Intn(50),
		HasSent:  false,
	}
}

func getSensor() (*Sensor, error) {
	sensor := NewSensor()
	return sensor, nil
}

func insertSensorData(ctx context.Context, client *mongo.Client, sensor Sensor) error {
	collection := client.Database(database).Collection(collection)
	fmt.Printf("data: %v\n", sensor)
	_, err := collection.InsertOne(ctx, sensor)
	if err != nil {
		return fmt.Errorf("failed to insert sensor data: %v", err)
	}
	return nil
}

func getCount(ctx context.Context, client *mongo.Client) (int64, error) {
	collection := client.Database(database).Collection(collection)
	filter := bson.D{{}}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count collection: %v", err)
	}
	return count, nil
}

func findSensor(ctx context.Context, client *mongo.Client) ([]bson.Raw, error) {
	collection := client.Database(database).Collection(collection)
	filter := bson.D{{Key: "HasSent", Value: "false"}}
	findOpt := options.Find()

	doc, err := collection.Find(ctx, filter, findOpt)
	fmt.Printf("doc: %v\n", doc)
	if err != nil {
		return nil, fmt.Errorf("failed to count collection: %v", err)
	}
	var sensors []bson.Raw
	if err = doc.All(ctx, &sensors); err != nil {
		return nil, fmt.Errorf("failed to decode sensor data: %v", err)
	}
	return sensors, nil
}
