package main

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"log"
	"strings"
	"time"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

type Item struct {
	ID   int64  `bson:"_id"`
	Data string `bson:"data"`
}

type ItemBuilder struct {
	Data string
}

func NewItemBuilder(size int) *ItemBuilder {
	sb := strings.Builder{}
	for i := 0; i < size; i++ {
		sb.WriteByte(letters[i%len(letters)])
	}
	return &ItemBuilder{Data: sb.String()}
}

func (ib *ItemBuilder) Build(id int64) *Item {
	return &Item{ID: id, Data: ib.Data}
}

func NewMongoClient() (*mongo.Client, func(), error) {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, err
	}
	return client, func() {
		_ = client.Disconnect(context.Background())
	}, nil
}

func benchmarkUpsert(client *mongo.Client, size int, batchCount int, total int) {
	db := client.Database("test")
	if err := db.Drop(context.Background()); err != nil {
		panic(err)
	}

	c := db.Collection("test")

	ib := NewItemBuilder(size)
	writeModels := make([]mongo.WriteModel, 0, batchCount)
	ts := time.Now()
	for i := 0; i < total; i += batchCount {
		writeModels = writeModels[:0]

		n := batchCount
		if i+n > total {
			n = total - i
		}

		for j := 0; j < n; j++ {
			item := ib.Build(int64(i + j))
			m := mongo.NewReplaceOneModel().SetFilter(bson.M{"_id": item.ID}).SetReplacement(item).SetUpsert(true)
			writeModels = append(writeModels, m)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			_, err := c.BulkWrite(ctx, writeModels)
			if err != nil {
				panic(err)
			}
			cancel()
		}
	}
	elapsed := time.Since(ts)
	log.Printf("Upsert %d items, size %d, batch %d, elapsed %v, qps %v\n", total, size, batchCount, elapsed, float64(total)/elapsed.Seconds())
}

func main() {
	client, cleanup, err := NewMongoClient()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	benchmarkUpsert(client, 4096, 100, 1000)
}
