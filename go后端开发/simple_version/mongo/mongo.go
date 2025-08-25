package mongo

import (
	"context"
	"os"
	"sync"
	"time"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	initOnce   sync.Once
	clientErr  error
)

func ensureClient(ctx context.Context) (*mongo.Client, error) {
	initOnce.Do(func() {
		uri := os.Getenv("MONGO_URI")
		if uri == "" { uri = "mongodb://localhost:27017" }
		c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil { log.Printf("MongoDB connect error: %v", err); clientErr = err; return }
		if err := c.Ping(ctx, nil); err != nil { log.Printf("MongoDB ping error: %v", err); clientErr = err; return }
		client = c
		log.Printf("MongoDB connected")
	})
	return client, clientErr
}

func getColl(ctx context.Context, name string) (*mongo.Collection, error) {
	c, err := ensureClient(ctx)
	if err != nil { log.Printf("MongoDB client unavailable: %v", err); return nil, err }
	db := os.Getenv("MONGO_DATABASE")
	if db == "" { db = "vnf_config" }
	return c.Database(db).Collection(name), nil
}

// YAMLReadDoc 首次/常规读取的快照文档
type YAMLReadDoc struct {
	ID        interface{}            `bson:"_id,omitempty"`
	Filename  string                 `bson:"filename"`
	ReadAt    time.Time              `bson:"read_at"`
	Content   interface{}            `bson:"content"`
	Fields    []map[string]interface{} `bson:"fields"`
}

// YAMLUpdateDoc 修改记录文档
type YAMLUpdateDoc struct {
	ID        interface{}            `bson:"_id,omitempty"`
	Filename  string                 `bson:"filename"`
	UpdatedAt time.Time              `bson:"updated_at"`
	Updates   map[string]interface{} `bson:"updates"`
	Content   interface{}            `bson:"content"`
	Fields    []map[string]interface{} `bson:"fields"`
}

// SaveYAMLRead 保存读取快照
func SaveYAMLRead(ctx context.Context, filename string, content interface{}, fields []map[string]interface{}) error {
	coll, err := getColl(ctx, "yaml_reads")
	if err != nil { return err }
	doc := YAMLReadDoc{Filename: filename, ReadAt: time.Now(), Content: content, Fields: fields}
	_, err = coll.InsertOne(ctx, doc)
	return err
}

// SaveYAMLUpdate 保存更新快照
func SaveYAMLUpdate(ctx context.Context, filename string, updates map[string]interface{}, content interface{}, fields []map[string]interface{}) error {
	coll, err := getColl(ctx, "yaml_updates")
	if err != nil { return err }
	doc := YAMLUpdateDoc{Filename: filename, UpdatedAt: time.Now(), Updates: updates, Content: content, Fields: fields}
	_, err = coll.InsertOne(ctx, doc)
	return err
}

// UpsertLatest 按文件维护一份最新快照（可选）
func UpsertLatest(ctx context.Context, filename string, content interface{}, fields []map[string]interface{}) error {
	coll, err := getColl(ctx, "yaml_latest")
	if err != nil { return err }
	_, err = coll.UpdateOne(ctx,
		bson.M{"filename": filename},
		bson.M{"$set": bson.M{"content": content, "fields": fields, "updated_at": time.Now()}},
		options.Update().SetUpsert(true),
	)
	return err
}
