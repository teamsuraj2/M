package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

const chatsIndex = "scChats"

func loadChatCache() ([]int64, error) {
	if val, ok := config.Cache.Load("chats"); ok {
		return val.([]int64), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var doc struct {
		ChatIDs []int64 `bson:"chats"`
	}

	err := chatDB.FindOne(ctx, bson.M{"_id": chatsIndex}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []int64{}, nil
		}
		return nil, err
	}

	config.Cache.Store("chats", doc.ChatIDs)
	return doc.ChatIDs, nil
}

func GetServedChats() ([]int64, error) {
	return loadChatCache()
}

func IsServedChat(chatID int64) (bool, error) {
	chats, err := loadChatCache()
	if err != nil {
		return false, err
	}
	for _, id := range chats {
		if id == chatID {
			return true, nil
		}
	}
	return false, nil
}

func AddServedChat(chatID int64) error {
	exists, err := IsServedChat(chatID)
	if err != nil || exists {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = chatDB.UpdateOne(ctx,
		bson.M{"_id": chatsIndex},
		bson.M{"$addToSet": bson.M{"chats": chatID}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	val, _ := config.Cache.LoadOrStore("chats", []int64{})
	chats := val.([]int64)
	chats = append(chats, chatID)
	config.Cache.Store("chats", chats)

	return nil
}

func DeleteServedChat(chatID int64) error {
	exists, err := IsServedChat(chatID)
	if err != nil || !exists {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = chatDB.UpdateOne(ctx,
		bson.M{"_id": chatsIndex},
		bson.M{"$pull": bson.M{"chats": chatID}},
	)
	if err != nil {
		return err
	}

	if val, ok := config.Cache.Load("chats"); ok {
		chats := val.([]int64)
		for i, id := range chats {
			if id == chatID {
				chats = append(chats[:i], chats[i+1:]...)
				break
			}
		}
		config.Cache.Store("chats", chats)
	}

	return nil
}

func MigrateChats() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cursor, err := chatDB.Find(ctx, bson.M{"chat_id": bson.M{"$lt": 0}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var oldChats []struct {
		ChatID int64 `bson:"chat_id"`
	}
	if err := cursor.All(ctx, &oldChats); err != nil {
		return err
	}

	if len(oldChats) == 0 {
		return nil
	}

	chatIDs := make([]int64, 0, len(oldChats))
	for _, chat := range oldChats {
		chatIDs = append(chatIDs, chat.ChatID)
	}

	_, err = chatDB.UpdateOne(ctx,
		bson.M{"_id": chatsIndex},
		bson.M{"$addToSet": bson.M{"chats": bson.M{"$each": chatIDs}}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	_, err = chatDB.DeleteMany(ctx, bson.M{
		"chat_id": bson.M{"$in": chatIDs},
	})
	if err != nil {
		return err
	}

	return nil
}