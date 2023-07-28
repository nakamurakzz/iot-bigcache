package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

const uri = "mongodb://root:example@mongo:27017"

func main() {
	if run() != 0 {
		fmt.Println("failed")
		return
	}
}

func run() int {
	const (
		success = iota
		failure
	)

	opt := options.Client().ApplyURI(uri).SetConnectTimeout(5 * time.Second)
	ctx := context.Background()

	fmt.Printf("connect to mongo: %s\n", uri)

	mongoClient, err := mongo.Connect(ctx, opt)
	if err != nil {
		fmt.Printf("failed to connect to mongo: %v", err)
		return failure
	}
	defer mongoClient.Disconnect(ctx)

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		fmt.Printf("failed to ping mongo: %v", err)
		return failure
	}

	fmt.Println("Connect Success.")

	database := mongoClient.Database("Test")
	_, err = database.Collection("Collection").InsertOne(ctx, map[string]string{"key": "value"})
	if err != nil {
		fmt.Printf("failed to insert: %v", err)
		return failure
	}

	errgroup, ctx := errgroup.WithContext(ctx)
	errgroup.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				fmt.Println("Hello")
				time.Sleep(1 * time.Second)
			}
		}
	})
	err = errgroup.Wait()
	if err != nil {
		fmt.Printf("failed to wait: %v", err)
		return failure
	}

	return success
}
