package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Entity struct {
	Id          uint64
	Name        string
	Description string
}

const dbName string = "poc2"

var coll *mongo.Collection
var seedData []Entity = []Entity{
	Entity{
		Id:          0,
		Name:        "First",
		Description: "the first element.",
	},
	Entity{
		Id:          1,
		Name:        "Second",
		Description: "the second element.",
	},
	Entity{
		Id:          2,
		Name:        "Third",
		Description: "the third element.",
	},
	Entity{
		Id:          3,
		Name:        "Fourth",
		Description: "the fourth element.",
	},
}
var dataBase []Entity = []Entity{
	Entity{
		Id:          0,
		Name:        "First",
		Description: "the first element.",
	},
	Entity{
		Id:          1,
		Name:        "Second",
		Description: "the second element.",
	},
	Entity{
		Id:          2,
		Name:        "Third",
		Description: "the third element.",
	},
	Entity{
		Id:          3,
		Name:        "Fourth",
		Description: "the fourth element.",
	},
}

func init() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		panic("MONGODB_URI missing")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	listDatabasesResult, err := client.ListDatabases(context.TODO(), nil)
	var exists bool = false
	for _, v := range listDatabasesResult.Databases {
		if v.Name == dbName {
			exists = true
		}
	}
	if exists {
		return
	}
	database := client.Database(dbName)
	coll = database.Collection("entities")
	coll.InsertMany(context.TODO(), seedData) // res, err :=
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
	cursor, err := coll.Find(context.TODO(), nil /*bson.D{{"name", "Bob"}} filter*/, opts)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	count, err := coll.CountDocuments(context.TODO(), nil)
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	send := struct {
		Results any
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
	var id uint64
	var err error
	var jsonByteArr []byte = []byte{}

	id, err = strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	var result bson.M
	err = coll.FindOne(
		context.TODO(),
		//bson.D{{"_id", id}},
		bson.D{{"Id", id}},
		nil,
	).Decode(&result)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonByteArr, err = json.Marshal(result)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(jsonByteArr) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	_, err = rw.Write(jsonByteArr)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

func RemoveEntity(rw http.ResponseWriter, req *http.Request) {
	var idParam string = req.PathValue("id")
	var id uint64
	var err error
	var isDeleted bool = false

	id, err = strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	for i, v := range dataBase {
		if v.Id == id {
			dataBase[i] = dataBase[len(dataBase)-1]
			dataBase = dataBase[:len(dataBase)-1]
			isDeleted = true
			break
		}
	}
	if isDeleted {
		rw.WriteHeader(http.StatusNoContent)
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}

func AddEntity(rw http.ResponseWriter, req *http.Request) {
	var e Entity
	var tmp map[string]any
	var id uint64 = 0

	err := json.NewDecoder(req.Body).Decode(&tmp) //decode to map/any instead of struct to validate not extra keys...
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
	//generate id and append
	for _, v := range dataBase {
		if v.Id > id {
			id = v.Id
		}
	}
	e.Id = id + 1
	dataBase = append(dataBase, e)
	rw.WriteHeader(http.StatusCreated)
	jsonByteArr, err := json.Marshal(e)
	if err != nil {
		return
	}
	rw.Write(jsonByteArr)
}

func UpdateEntity(rw http.ResponseWriter, req *http.Request) {
	var idParam string = req.PathValue("id")
	var id uint64
	var err error
	var isUpdated bool = false
	id, err = strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	var tmp map[string]any

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
	for i, v := range dataBase {
		if v.Id == id {
			dataBase[i].Name = valName.(string)
			dataBase[i].Description = valDescription.(string)
			isUpdated = true
			break
		}
	}
	if isUpdated {
		rw.WriteHeader(http.StatusNoContent)
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}
