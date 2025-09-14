package database

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

func SetNSFWFlag(chatID int64, enable bool) error {
	cacheKey := fmt.Sprintf("%d_nsfw", chatID)

	if cached, ok := config.Cache.Load(cacheKey); ok {
		if cached.(bool) == enable {
			return nil
		}
	}

	old := IsNSFWEnabled(chatID)

	if old == enable {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := nsfwFlagsDB.UpdateOne(
		ctx,
		bson.M{"_id": chatID},
		bson.M{"$set": bson.M{"enabled": enable}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		config.Cache.Store(cacheKey, enable)
	}
	return err
}

func IsNSFWEnabled(chatID int64) bool {
	key := fmt.Sprintf("%d_nsfw", chatID)

	if val, ok := config.Cache.Load(key); ok {
		if enabled, ok := val.(bool); ok {
			return enabled
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result struct {
		Enabled bool `bson:"enabled"`
	}

	err := nsfwFlagsDB.FindOne(ctx, bson.M{"_id": chatID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			config.Cache.Store(key, false)
			return false
		}
		fmt.Println("Database error IsNsfwOn", err)
		return false
	}

	config.Cache.Store(key, result.Enabled)
	return result.Enabled
}
