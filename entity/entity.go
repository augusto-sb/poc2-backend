package entity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

    "go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Entity struct {
	Id          bson.ObjectID `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

const dbName string = "poc2"
const collectionName string = "entities"

var coll *mongo.Collection
var seedData []Entity = []Entity{
	Entity{
		Name:        "First",
		Description: "the first element.",
	},
	Entity{
		Name:        "Second",
		Description: "the second element.",
	},
	Entity{
		Name:        "Third",
		Description: "the third element.",
	},
	Entity{
		Name:        "Fourth",
		Description: "the fourth element.",
	},
}

func gracefulShutdown(client *mongo.Client) {
    s := make(chan os.Signal, 1)
    signal.Notify(s, os.Interrupt)
    signal.Notify(s, syscall.SIGTERM)
    go func() {
        <-s
        fmt.Println("Sutting down gracefully.")
        err := client.Disconnect(context.TODO());
		if err != nil {
			//panic(err)
		}
        os.Exit(0)
    }()
}

func init() {
	var err error
	var res *mongo.InsertManyResult
	// https://pkg.go.dev/go.mongodb.org/mongo-driver/v2
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		panic("MONGODB_URI missing")
	}
	opts := options.Client().ApplyURI(uri).SetTimeout(5 * time.Second)
	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}
	ping := client.Ping(context.TODO(), readpref.Primary())
	fmt.Println(ping)
	go gracefulShutdown(client)
	listDatabasesResult, err := client.ListDatabases(context.TODO(), bson.M{})
	var exists bool = false
	for _, v := range listDatabasesResult.Databases {
		if v.Name == dbName {
			exists = true
			break
		}
	}
	database := client.Database(dbName)
	coll = database.Collection(collectionName)
	if exists {
		return
	}
	asdasd := options.IndexOptionsBuilder{}
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
		Options: asdasd.SetUnique(true),
	}
	coll.Indexes().CreateOne(context.TODO(), index)
	res, err = coll.InsertMany(context.TODO(), seedData)
	if(err == nil){
		fmt.Println(res)
	}else{
		fmt.Println("error inserting seed data", err)
	}
}

func GetEntities(rw http.ResponseWriter, req *http.Request) {
	var page uint64 = 0
	var size uint64 = 2
	var sortKey string = ""
	var sortDirection bool = true
	var err error

	query := req.URL.Query()
	querySortKeyVal, querySortKeyOk := query["sort.key"]
	if querySortKeyOk && len(querySortKeyVal) == 1 {
		if querySortKeyVal[0] == "Id" || querySortKeyVal[0] == "Name" || querySortKeyVal[0] == "Description" {
			sortKey = querySortKeyVal[0]
		}
	}
	querySortDirectionVal, querySortDirectionOk := query["sort.direction"]
	if querySortDirectionOk && len(querySortDirectionVal) == 1 {
		if querySortDirectionVal[0] == "false" {
			sortDirection = false
		}
	}
	queryPageVal, queryPageOk := query["page"]
	if queryPageOk && len(queryPageVal) == 1 {
		page, err = strconv.ParseUint(queryPageVal[0], 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("page should be numeric"))
			return
		}
	}
	querySizeVal, querySizeOk := query["size"]
	if querySizeOk && len(querySizeVal) == 1 {
		size, err = strconv.ParseUint(querySizeVal[0], 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("size should be numeric"))
			return
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(size))
	opts.SetSkip(int64(page * size))
	if sortKey != "" {
		if sortDirection {
			opts.SetSort(bson.D{{sortKey, 1}})
		} else {
			opts.SetSort(bson.D{{sortKey, -1}})
		}
	}
	filter := bson.D{}
	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	var results []Entity
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	send := struct {
		Results []Entity
		Count   int64
	}{Results: results, Count: count}
	jsonByteArr, err := json.Marshal(send)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = rw.Write(jsonByteArr)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

func GetEntity(rw http.ResponseWriter, req *http.Request) {
	var idParam string = req.PathValue("id")
	var err error
	var jsonByteArr []byte = []byte{}
	var objId bson.ObjectID
	var result Entity

	objId, err = bson.ObjectIDFromHex(idParam)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = coll.FindOne(
		context.TODO(),
		bson.M{"_id": objId},
		nil,
	).Decode(&result)
	if err != nil {
		if(errors.Is(err, mongo.ErrNoDocuments)){
			rw.WriteHeader(http.StatusNotFound)
		}else{
			rw.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	jsonByteArr, err = json.Marshal(result)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = rw.Write(jsonByteArr)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

func RemoveEntity(rw http.ResponseWriter, req *http.Request) {
	var idParam string = req.PathValue("id")
	var err error
	var result *mongo.DeleteResult
	var objId bson.ObjectID

	objId, err = bson.ObjectIDFromHex(idParam)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err = coll.DeleteOne(context.TODO(), bson.M{"_id": objId})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 1 {
		rw.WriteHeader(http.StatusNoContent)
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}

func AddEntity(rw http.ResponseWriter, req *http.Request) {
	var e Entity
	var tmp map[string]any
	var insertOneResult *mongo.InsertOneResult
	var err error
	var jsonByteArr []byte

	err = json.NewDecoder(req.Body).Decode(&tmp) //decode to map/any instead of struct to validate not extra keys...
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	//validations
	if len(tmp) != 2 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("expected exactly 2 keys"))
		return
	}
	valName, okName := tmp["Name"]
	valDescription, okDescription := tmp["Description"]
	if !okName || !okDescription {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Name or Description missing!"))
		return
	}
	switch valName.(type) {
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Name should be string!"))
		return
	case string:
		e.Name = valName.(string)
	}
	switch valDescription.(type) {
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Description should be string!"))
		return
	case string:
		e.Description = valDescription.(string)
	}

	insertOneResult, err = coll.InsertOne(context.TODO(), e, nil);
	if err != nil {
		if(mongo.IsDuplicateKeyError(err)){
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("ese Name ya está en uso!"))
		}else{
			rw.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	e.Id = insertOneResult.InsertedID.(bson.ObjectID)

	rw.WriteHeader(http.StatusCreated)
	jsonByteArr, err = json.Marshal(e)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Write(jsonByteArr)
}

func UpdateEntity(rw http.ResponseWriter, req *http.Request) {
	var idParam string = req.PathValue("id")
	var err error
	var tmp map[string]any
	var objId bson.ObjectID

	objId, err = bson.ObjectIDFromHex(idParam)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(req.Body).Decode(&tmp) //decode to map/any instead of struct to validate not extra keys...
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	//validations
	if len(tmp) != 2 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("expected exactly 2 keys"))
		return
	}
	valName, okName := tmp["Name"]
	valDescription, okDescription := tmp["Description"]
	if !okName || !okDescription {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Name or Description missing!"))
		return
	}
	switch valName.(type) {
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Name should be string!"))
		return
	case string:
		//is ok!
	}
	switch valDescription.(type) {
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Description should be string!"))
		return
	case string:
		//is ok!
	}
	//find and update
	opts := options.UpdateOne().SetUpsert(false)
	filter := bson.D{{"_id", objId}}
	update := bson.D{{"$set", bson.D{{"Name", valName.(string)},{"Description", valDescription.(string)}}}}
	result, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		fmt.Println(err)
	}
	if result.MatchedCount != 0 {
		rw.WriteHeader(http.StatusNoContent)
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}
