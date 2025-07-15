package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

const usersIndex = "scUsers"

func loadUserCache() ([]int64, error) {
	if val, ok := config.Cache.Load("users"); ok {
		return val.([]int64), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var doc struct {
		UserIDs []int64 `bson:"users"`
	}

	err := userDB.FindOne(ctx, bson.M{"_id": usersIndex}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []int64{}, nil
		}
		return nil, err
	}

	config.Cache.Store("users", doc.UserIDs)
	return doc.UserIDs, nil
}

func GetServedUsers() ([]int64, error) {
	return loadUserCache()
}

func IsServedUser(userID int64) (bool, error) {
	users, err := loadUserCache()
	if err != nil {
		return false, err
	}
	for _, id := range users {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
}

func AddServedUser(userID int64) error {
	exists, err := IsServedUser(userID)
	if err != nil || exists {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = userDB.UpdateOne(ctx,
		bson.M{"_id": usersIndex},
		bson.M{"$addToSet": bson.M{"users": userID}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	val, _ := config.Cache.LoadOrStore("users", []int64{})
	users := val.([]int64)
	users = append(users, userID)
	config.Cache.Store("users", users)

	return nil
}

func DeleteServedUser(userID int64) error {
	exists, err := IsServedUser(userID)
	if err != nil || !exists {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = userDB.UpdateOne(ctx,
		bson.M{"_id": usersIndex},
		bson.M{"$pull": bson.M{"users": userID}},
	)
	if err != nil {
		return err
	}

	if val, ok := config.Cache.Load("users"); ok {
		users := val.([]int64)
		for i, id := range users {
			if id == userID {
				users = append(users[:i], users[i+1:]...)
				break
			}
		}
		config.Cache.Store("users", users)
	}

	return nil
}

func MigrateUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cursor, err := userDB.Find(ctx, bson.M{"user_id": bson.M{"$gt": 0}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var oldUsers []struct {
		UserID int64 `bson:"user_id"`
	}
	if err := cursor.All(ctx, &oldUsers); err != nil {
		return err
	}

	if len(oldUsers) == 0 {
		return nil
	}

	userIDs := make([]int64, 0, len(oldUsers))
	for _, user := range oldUsers {
		userIDs = append(userIDs, user.UserID)
	}

	_, err = userDB.UpdateOne(ctx,
		bson.M{"_id": usersIndex},
		bson.M{"$addToSet": bson.M{"users": bson.M{"$each": userIDs}}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	_, err = userDB.DeleteMany(ctx, bson.M{
		"user_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		return err
	}

	return nil
}