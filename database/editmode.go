package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

var (
	defaultEditModeDuration = 0
	defaultEditMode         = "USER"
)

// mode for a chat ("ADMIN", "USER", "OFF").
type EditModeSettings struct {
	ChatID   int64  `bson:"chat_id"`
	Mode     string `bson:"mode"`
	Duration int    `bson:"duration"`
}

func SetEditMode(setting EditModeSettings) (bool, error) {
	key := fmt.Sprintf("editmode:%d", setting.ChatID)

	if cachedVal, ok := config.Cache.Load(key); ok {
		if cachedSetting, valid := cachedVal.(EditModeSettings); valid {
			if cachedSetting.Mode == setting.Mode && cachedSetting.Duration == setting.Duration {
				return true, nil
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"mode":     setting.Mode,
			"duration": setting.Duration,
		},
	}

	result, err := editModeDB.UpdateOne(
		ctx,
		bson.M{"chat_id": setting.ChatID},
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		log.Printf("SetEditMode error for chatID %d: %v", setting.ChatID, err)
		return false, err
	}

	config.Cache.Store(key, setting)
	return result.ModifiedCount > 0 || result.UpsertedCount > 0, nil
}

func GetEditMode(chatID int64) EditModeSettings {
	key := fmt.Sprintf("editmode:%d", chatID)

	if cachedVal, ok := config.Cache.Load(key); ok {
		if settings, valid := cachedVal.(EditModeSettings); valid {
			return settings
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var settings EditModeSettings
	err := editModeDB.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&settings)
	if err != nil || settings.Mode == "" {
		settings = EditModeSettings{
			ChatID:   chatID,
			Mode:     defaultEditMode,
			Duration: defaultEditModeDuration,
		}
	}

	config.Cache.Store(key, settings)
	return settings
}

func ResetEditMode(chatID int64) error {
	setting := EditModeSettings{
		ChatID:   chatID,
		Mode:     defaultEditMode,
		Duration: defaultEditModeDuration,
	}
	_, err := SetEditMode(setting)
	return err
}
