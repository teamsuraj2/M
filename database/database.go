package database

import (
	"context"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

var (
	client      *mongo.Client
	userDB      *mongo.Collection
	chatDB      *mongo.Collection
	editModeDB  *mongo.Collection
	echoDB      *mongo.Collection
	loggerDB    *mongo.Collection
	linksDB     *mongo.Collection
	bioLinkDB   *mongo.Collection
	nsfwWordsDB *mongo.Collection
	nsfwFlagsDB *mongo.Collection
	timeout     = 5 * time.Second
)

func init() {
	if config.MongoUri == "" {
		log.Panic("MongoDB URI is missing in config.MongoUri")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		log.Panicf("Failed to connect to MongoDB: %v", err)
	}

	db := client.Database("EditGuardainBot")
	userDB = db.Collection("userstats")
	chatDB = db.Collection("chats")
	editModeDB = db.Collection("editmodes")
	echoDB = db.Collection("echos")
	loggerDB = db.Collection("logger")
	linksDB = db.Collection("links")
	bioLinkDB = db.Collection("biolinks")
	nsfwWordsDB = db.Collection("nsfw_words")
	nsfwFlagsDB = db.Collection("nsfw_flags")

	// Indexes
	CreateIndex(ctx, userDB, bson.D{{Key: "user_id", Value: 1}}, true)
	CreateIndex(ctx, chatDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, editModeDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, echoDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, linksDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, loggerDB, bson.D{{Key: "enabled", Value: 1}}, false)
}

func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Error while disconnecting MongoDB: %v", err)
	}
}

func CreateIndex(ctx context.Context, coll *mongo.Collection, keys bson.D, unique bool) {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique),
	}
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		if strings.Contains(err.Error(), "IndexOptionsConflict") {
			// Drop all indexes and retry
			if errDrop := DropAllIndexes(ctx, coll); errDrop != nil {
				log.Printf("❌ Failed to drop indexes on %s: %v\n", coll.Name(), errDrop)
				return
			}
			// Retry creating the index
			_, err = coll.Indexes().CreateOne(ctx, indexModel)
		}
		if err != nil {
			log.Printf("❌ Failed to create index on %s: %v\n", coll.Name(), err)
			return
		}
	}
	log.Printf("✅ Index created on %s with keys: %v, unique: %v\n", coll.Name(), keys, unique)
}

func DropAllIndexes(ctx context.Context, coll *mongo.Collection) error {
	err := coll.Indexes().DropAll(ctx)
	return err
}
