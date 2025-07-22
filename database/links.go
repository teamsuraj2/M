package database

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LinkData struct {
	ChatID    int64    `bson:"chat_id"`
	Hostnames []string `bson:"hostnames"`
	Enabled   bool     `bson:"enabled"`
}

func GetAllowedHostnames(chatID int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var entry LinkData
	err := linksDB.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	return entry.Hostnames, nil
}

func AddAllowedHostname(chatID int64, hostname string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := linksDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$addToSet": bson.M{"hostnames": hostname}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func RemoveAllowedHostname(chatID int64, hostname string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := linksDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$pull": bson.M{"hostnames": hostname}},
	)
	return err
}

func IsLinkFilterEnabled(chatID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var entry LinkData
	err := linksDB.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return entry.Enabled, nil
}

func SetLinkFilterEnabled(chatID int64, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := linksDB.UpdateOne(ctx,
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"enabled": enabled}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}