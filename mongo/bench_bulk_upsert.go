package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	B  = 1
	KB = B << 10
	MB = KB << 10
	GB = MB << 10
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

type Item struct {
	ID   int64  `bson:"_id"`
	Data string `bson:"data"`
}

type ItemBuilder struct {
	Data string
}

func byteSize(size int) string {
	switch {
	case size > GB:
		return fmt.Sprintf("%.2fGB", float64(size/GB)+float64(size%GB)/float64(GB))
	case size > MB:
		return fmt.Sprintf("%.2fMB", float64(size/MB)+float64(size%MB)/float64(MB))
	case size > KB:
		return fmt.Sprintf("%.2fKB", float64(size/KB)+float64(size%KB)/float64(KB))
	default:
		return fmt.Sprintf("%dB", size)
	}
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

		n := batchCount
		if i+n > total {
			n = total - i
		}

		writeModels = writeModels[:0]
		for j := 0; j < n; j++ {
			item := ib.Build(int64(i + j))
			m := mongo.NewReplaceOneModel().SetFilter(bson.M{"_id": item.ID}).SetReplacement(item).SetUpsert(true)
			writeModels = append(writeModels, m)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
		_, err := c.BulkWrite(ctx, writeModels)
		if err != nil {
			panic(err)
		}
		cancel()
	}
	elapsed := time.Since(ts)

	write("bulkWrite/upsert", byteSize(size), batchCount, byteSize(batchCount*size), total, byteSize(size*total), elapsed, elapsed/time.Duration(total/batchCount), float64(total)/elapsed.Seconds(), fmt.Sprintf("%s/s", byteSize(int(float64(size*total)/elapsed.Seconds()))))
}

func write(row ...any) {
	elems := make([]string, 0, len(row))
	for _, v := range row {
		switch t := v.(type) {
		case float64, float32:
			elems = append(elems, fmt.Sprintf("%.2f", t))
		case time.Duration:
			if t > time.Second {
				elems = append(elems, fmt.Sprintf("%.2fs", t.Seconds()))
			} else if t > time.Millisecond {
				elems = append(elems, fmt.Sprintf("%dms", t.Milliseconds()))
			} else {
				elems = append(elems, fmt.Sprintf("%dus", t.Microseconds()))
			}
		default:
			elems = append(elems, fmt.Sprint(v))
		}
	}
	fmt.Println(strings.Join(elems, "\t"))
}

var (
	size       int
	batchCount int
	total      int
	head       bool
)

func main() {
	flag.IntVar(&size, "size", 4, "doc size")
	flag.IntVar(&batchCount, "batchCount", 1, "batchCount")
	flag.IntVar(&total, "total", 10000, "total")
	flag.BoolVar(&head, "head", false, "head")
	flag.Parse()

	if head {
		write("op", "size", "batch", "batchSize", "total", "totalSize", "time", "batchTime", "qps", "writeSpeed")
	}

	client, cleanup, err := NewMongoClient()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	benchmarkUpsert(client, size*KB, batchCount, total)
}
