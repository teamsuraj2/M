package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

const defaultLogger bool = true

func IsLoggerEnabled() bool {
	const key = "logger"

	if val, ok := config.Cache.Load(key); ok {
		if enabled, valid := val.(bool); valid {
			return enabled
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result struct {
		Enabled bool `bson:"enabled"`
	}

	err := loggerDB.FindOne(ctx, bson.M{}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			config.Cache.Store(key, defaultLogger)
			return defaultLogger
		}
		log.Println(err)
		return defaultLogger
	}

	config.Cache.Store(key, result.Enabled)
	return result.Enabled
}

func SetLogger(enabled bool) error {
	const key = "logger"

	if val, ok := config.Cache.Load(key); ok {
		if cached, valid := val.(bool); valid && cached == enabled {
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	update := bson.M{"$set": bson.M{"enabled": enabled}}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := loggerDB.UpdateOne(ctx, bson.M{}, update, opts)
	if err == nil {
		config.Cache.Store(key, enabled)
	}
	return err
}
