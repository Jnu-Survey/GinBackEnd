package common

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	MongoPool *MongoDb
)

type MongoDb struct {
	connection *mongo.Collection
}

func NewMongoDbPool() (*MongoDb, error) {
	pool, err := ConnectToDB()
	if err != nil {
		return nil, err
	}
	return &MongoDb{
		connection: pool,
	}, nil
}

func ConnectToDB() (*mongo.Collection, error) {
	url := MongoDataSource
	name := MongoName
	collection := MongoCollection
	maxCollection := MongoMaxCollection
	var timeout time.Duration = 10 // 设置10秒的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	o := options.Client().ApplyURI(url)
	o.SetMaxPoolSize(uint64(maxCollection))
	client, err := mongo.Connect(ctx, o)
	if err != nil {
		return nil, err
	}
	return client.Database(name).Collection(collection), nil
}

func (m *MongoDb) jsonStr2Bson(str string) (interface{}, error) {
	var want interface{}
	err := bson.UnmarshalJSON([]byte(str), &want)
	if err != nil {
		return nil, err
	}
	return want, nil
}

func (m *MongoDb) InsertToDb(wantStr string) (string, error) {
	if wantStr == "" {
		return "", errors.New("转换的字符串为空")
	}
	want, err := m.jsonStr2Bson(wantStr)
	if err != nil {
		return "", err
	}
	res, err := m.connection.InsertOne(context.TODO(), want)
	if err != nil {
		return "", err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("断言错误")
	}
	return id.Hex(), nil
}

func (m *MongoDb) FindInfoByField(field, want string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	filter := bson.M{field: want}
	cursor, err := m.connection.Find(ctx, filter)
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)
	var temp []bson.M
	if err = cursor.All(context.Background(), &temp); err != nil {
		return "", err
	}
	if len(temp) == 0 {
		return "", nil
	}
	jsonInfo, err := json.Marshal(temp)
	if err != nil {
		return "", err
	}
	return string(jsonInfo), nil
}

func (m *MongoDb) FindInfoById(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	cursor, err := m.connection.Find(ctx, filter)
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)
	var temp []bson.M
	if err = cursor.All(context.Background(), &temp); err != nil {
		return "", err
	}
	if len(temp) == 0 {
		return "", nil
	}
	jsonInfo, err := json.Marshal(temp[0])
	if err != nil {
		return "", err
	}
	return string(jsonInfo), nil
}
