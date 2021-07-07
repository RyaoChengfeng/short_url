package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"src/config"
	. "src/util"
	"time"
)

type Url struct {
	ID       primitive.ObjectID `bson:"_id"`
	LongUrl  string             `bson:"long_url"`
	ShortUrl string             `bson:"short_url"`
}

type model struct {
	dbTrait
	ctx   context.Context
	abort bool
}

type Model interface {
	// Close 关闭数据库连接
	Close()
	// Abort 终止操作，用于如事务的取消
	Abort()
	InsertUrl(url Url) (*mongo.InsertOneResult, error)
	UpdateUrlWithID(url Url, id primitive.ObjectID) (*mongo.UpdateResult, error)
	RetrieveUrlWithShortUrl(shortUrl string) (Url, error)
}

func GetModel() Model {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	if config.C.Debug {
		ctx = context.Background()
	}
	ret := &model{
		dbTrait: getMongoDBTx(ctx),
		ctx:     ctx,
		abort:   false,
	}
	return ret
}

func (m *model) Close() {
	// DO NOTHING
}

func (m *model) Abort() {
	m.abort = true
}

var (
	mongoClient *mongo.Client
)

const (
	collectionName = "url"
)

type dbTrait struct {
	db *mongo.Database
}

func initMongoMongoDB() {
	var err error
	uri := fmt.Sprintf("mongodb://%s", config.C.MongoDB.Addr)
	mongoClient, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	// Ping test
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err = mongoClient.Connect(ctx)
	if err != nil {
		panic(err)
	}
	//defer cancel()

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}
	Logger.Println("Database init done!")
}

func GetMongoGlobalClient() *mongo.Client {
	return mongoClient
}

// session 事务，但是需要mongo在cluster模式，慎用
func session(ctx context.Context, f func(ctx mongo.SessionContext) error) error {
	session, err := mongoClient.StartSession()
	if err != nil {
		return err
	}

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	err = mongo.WithSession(ctx, session, f)

	if err != nil {
		err = session.AbortTransaction(ctx)
	} else {
		err = session.CommitTransaction(ctx)
	}

	session.EndSession(ctx)
	return err
}

func getMongoDBTx(ctx context.Context) dbTrait {
	err := mongoClient.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	return dbTrait{
		db: mongoClient.Database(config.C.MongoDB.DB),
	}
}

func mongoCreateDocument(collection string, data interface{}) (*mongo.InsertOneResult, error) {
	col := mongoClient.Database(config.C.MongoDB.DB).Collection(collection)
	doc, err := col.InsertOne(context.TODO(), data)
	return doc, err
}

func (m *model) InsertUrl(url Url) (*mongo.InsertOneResult, error) {
	return mongoCreateDocument(collectionName, url)
}

func (m *model) UpdateUrlWithID(url Url, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": id}
	col := mongoClient.Database(config.C.MongoDB.DB).Collection(collectionName)
	opts := options.Update().SetUpsert(false)
	update := bson.M{"$set": url}
	result, err := col.UpdateOne(context.TODO(), filter, update, opts)
	return result, err
}

func (m *model) RetrieveUrlWithShortUrl(shortUrl string) (Url, error) {
	var url Url
	filter := bson.M{"short_url": shortUrl}
	col := mongoClient.Database(config.C.MongoDB.DB).Collection(collectionName)
	err := col.FindOne(context.TODO(), filter).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Url{}, nil
		}
		return Url{}, err
	}
	return url, nil
}
