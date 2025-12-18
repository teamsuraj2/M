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

type MsgDeleteSettings struct {
	ChatID  int64         `bson:"chat_id"`
	Enabled bool          `bson:"enabled"`
	Delay   time.Duration `bson:"delay"`
}

func GetMsgDeleteSettings(chatID int64) (*MsgDeleteSettings, error) {
	key := fmt.Sprintf("msg_delete:%d", chatID)
	if val, ok := config.Cache.Load(key); ok {
		return val.(*MsgDeleteSettings), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var settings MsgDeleteSettings
	err := msgDeleteDB.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		settings = MsgDeleteSettings{
			ChatID:  chatID,
			Enabled: false,
			Delay:   1 * time.Hour,
		}
	} else if err != nil {
		return nil, err
	}

	config.Cache.Store(key, &settings)
	return &settings, nil
}

func SetMsgDeleteEnabled(chatID int64, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := msgDeleteDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"enabled": enabled}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		key := fmt.Sprintf("msg_delete:%d", chatID)
		config.Cache.Delete(key)
	}
	return err
}

func SetMsgDeleteDelay(chatID int64, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := msgDeleteDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"delay": delay}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		key := fmt.Sprintf("msg_delete:%d", chatID)
		config.Cache.Delete(key)
	}
	return err
}
