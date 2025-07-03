package database

import (
        "context"

        "go.mongodb.org/mongo-driver/v2/bson"
        "go.mongodb.org/mongo-driver/v2/mongo"
        "go.mongodb.org/mongo-driver/v2/mongo/options"
)

const bioModeDocID = "bio_mode"

func SetBioMode(chatID int64) error {
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()

        _, err := bioLinkDB.UpdateOne(
                ctx,
                bson.M{"_id": bioModeDocID},
                bson.M{"$addToSet": bson.M{"chats": chatID}}, // add only if not exists
                options.UpdateOne().SetUpsert(true),
        )
        return err
}

func GetBioMode(chatID int64) (bool, error) {
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()

        var result struct {
                Chats []int64 `bson:"chats"`
        }

        err := bioLinkDB.FindOne(ctx, bson.M{"_id": bioModeDocID}).Decode(&result)
        if err != nil {
                if err == mongo.ErrNoDocuments {
                        return false, nil
                }
                return false, err
        }

        for _, id := range result.Chats {
                if id == chatID {
                        return true, nil
                }
        }
        return false, nil
}

func DelBioMode(chatID int64) error {
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()

        _, err := bioLinkDB.UpdateOne(
                ctx,
                bson.M{"_id": bioModeDocID},
                bson.M{"$pull": bson.M{"chats": chatID}}, // remove from array
        )
        return err
}