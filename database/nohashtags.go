package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

func IsNoHashtagsEnabled(chatID int64) (bool, error) {
	key := fmt.Sprintf("no_hashtags:%d", chatID)
	if val, ok := config.Cache.Load(key); ok {
		return val.(bool), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result struct {
		Enabled bool `bson:"enabled"`
	}
	err := noHashtagsDB.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		config.Cache.Store(key, false)
		return false, nil
	} else if err != nil {
		return false, err
	}

	config.Cache.Store(key, result.Enabled)
	return result.Enabled, nil
}

func SetNoHashtagsEnabled(chatID int64, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := noHashtagsDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"enabled": enabled}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		key := fmt.Sprintf("no_hashtags:%d", chatID)
		config.Cache.Store(key, enabled)
	}
	return err
}
