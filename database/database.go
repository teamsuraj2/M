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
	nsfwFlagsDB *mongo.Collection

	mediaDeleteDB, fileDeleteDB, msgDeleteDB *mongo.Collection
	noForwardDB, noPhoneDB,  noHashtagsDB, noPromoDB                    *mongo.Collection
	timeout                                  = 11 * time.Second
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
	nsfwFlagsDB = db.Collection("nsfw_flags")
	mediaDeleteDB = db.Collection("media_delete")
	fileDeleteDB = db.Collection("file_delete")
	msgDeleteDB = db.Collection("msg_delete")
	noForwardDB = db.Collection("no_forward")
	noPhoneDB = db.Collection("no_phone")
	noHashtagsDB = db.Collection("no_hashtags")
	noPromoDB = db.Collection("no_promo")

	// Indexes

	CreateIndex(ctx, editModeDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, echoDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, linksDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, loggerDB, bson.D{{Key: "enabled", Value: 1}}, false)
	CreateIndex(ctx, mediaDeleteDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, fileDeleteDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, msgDeleteDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, noForwardDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, noPhoneDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, noHashtagsDB, bson.D{{Key: "chat_id", Value: 1}}, true)
	CreateIndex(ctx, noPromoDB, bson.D{{Key: "chat_id", Value: 1}}, true)
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
}

func DropAllIndexes(ctx context.Context, coll *mongo.Collection) error {
	err := coll.Indexes().DropAll(ctx)
	return err
}
