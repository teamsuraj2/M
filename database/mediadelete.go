package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

type MediaDeleteSettings struct {
	ChatID           int64         `bson:"chat_id"`
	Enabled          bool          `bson:"enabled"`
	Delay            time.Duration `bson:"delay"`
	DeleteFromUsers  bool          `bson:"delete_from_users"`  // Delete from regular users
	DeleteFromAdmins bool          `bson:"delete_from_admins"` // Delete from admins too
}

func GetMediaDeleteSettings(chatID int64) (*MediaDeleteSettings, error) {
	key := fmt.Sprintf("media_delete:%d", chatID)
	if val, ok := config.Cache.Load(key); ok {
		return val.(*MediaDeleteSettings), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var settings MediaDeleteSettings
	err := mediaDeleteDB.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		settings = MediaDeleteSettings{
			ChatID:           chatID,
			Enabled:          false,
			Delay:            6 * time.Hour,
			DeleteFromUsers:  true,
			DeleteFromAdmins: false,
		}
	} else if err != nil {
		return nil, err
	}

	config.Cache.Store(key, &settings)
	return &settings, nil
}

func SetMediaDeleteEnabled(chatID int64, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := mediaDeleteDB.UpdateOne(
		ctx,
		bson.M{"chat_id": chatID},
		bson.M{
			"$set": bson.M{
				"enabled": enabled,
			},
			"$setOnInsert": bson.M{
				"delay":              6 * time.Hour,
				"delete_from_users":  true,
				"delete_from_admins": false,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)

	if err == nil {
		key := fmt.Sprintf("media_delete:%d", chatID)
		config.Cache.Delete(key)
	}
	return err
}

func SetMediaDeleteDelay(chatID int64, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := mediaDeleteDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"delay": delay}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		key := fmt.Sprintf("media_delete:%d", chatID)
		config.Cache.Delete(key)
	}
	return err
}

func SetMediaDeletionMode(chatID int64, deleteFromUsers, deleteFromAdmins bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := mediaDeleteDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{
			"delete_from_users":  deleteFromUsers,
			"delete_from_admins": deleteFromAdmins,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		key := fmt.Sprintf("media_delete:%d", chatID)
		config.Cache.Delete(key)
	}
	return err
}
