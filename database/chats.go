package database

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"main/config"
)

func GetServedChats() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cursor, err := chatDB.Find(ctx, bson.M{"chat_id": bson.M{"$lt": 0}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chats []struct {
		ChatID int64 `bson:"chat_id"`
	}
	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}

	var chatIDs []int64
	for _, chat := range chats {
		chatIDs = append(chatIDs, chat.ChatID)
	}

	return chatIDs, nil
}

func IsServedChat(chatID int64) (bool, error) {
	if val, ok := config.Cache.Load("chats"); ok {
		chats := val.([]int64)
		for _, id := range chats {
			if id == chatID {
				return true, nil
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	count, err := chatDB.CountDocuments(ctx, bson.M{"chat_id": chatID})
	if err != nil {
		return false, err
	}

	if count > 0 {
		if val, ok := config.Cache.Load("chats"); ok {
			chats := val.([]int64)
			config.Cache.Store("chats", append(chats, chatID))
		} else {
			config.Cache.Store("chats", []int64{chatID})
		}
	}

	return count > 0, nil
}

func AddServedChat(chatID int64) error {
	exists, err := IsServedChat(chatID)
	if err != nil || exists {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		_, err := chatDB.InsertOne(ctx, bson.M{"chat_id": chatID})
		if err == nil {
			if val, ok := config.Cache.Load("chats"); ok {
				chats := val.([]int64)
				config.Cache.Store("chats", append(chats, chatID))
			} else {
				config.Cache.Store("chats", []int64{chatID})
			}
		}
	}()

	return nil
}

func DeleteServedChat(chatID int64) error {
	exists, err := IsServedChat(chatID)
	if err != nil || !exists {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		_, err := chatDB.DeleteOne(ctx, bson.M{"chat_id": chatID})
		if err == nil {
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
		}
	}()

	return nil
}
