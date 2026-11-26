package backup

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dendianugerah/velld/internal/connection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m branchMasker) mongo(ctx context.Context, conn *connection.StoredConnection, user, password string) error {
	client, err := openMongoTarget(ctx, conn)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())
	db := client.Database(conn.DatabaseName)
	collections, err := db.ListCollectionSpecifications(ctx, bson.D{})
	if err != nil {
		return err
	}
	for _, spec := range collections {
		if spec.Type == "view" || strings.HasPrefix(spec.Name, "system.") {
			continue
		}
		if err := m.maskMongoCollection(ctx, db.Collection(spec.Name)); err != nil {
			return err
		}
	}
	return db.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: user},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: "readWrite"}, {Key: "db", Value: conn.DatabaseName}}}},
	}).Err()
}

func (m branchMasker) maskMongoCollection(ctx context.Context, collection *mongo.Collection) error {
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc bson.D
		if err := cursor.Decode(&doc); err != nil {
			return err
		}
		var id any
		for _, entry := range doc {
			if entry.Key == "_id" {
				id = entry.Value
				break
			}
		}
		value, changed, err := m.maskBSON(doc, "")
		if err != nil {
			return err
		}
		if changed {
			filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$eq", Value: id}}}}
			result, err := collection.ReplaceOne(ctx, filter, value)
			if err != nil {
				return err
			}
			if result.MatchedCount != 1 {
				return fmt.Errorf("sensitive document could not be updated")
			}
		}
	}
	return cursor.Err()
}

func (m branchMasker) maskBSON(value any, kind piiKind) (any, bool, error) {
	switch v := value.(type) {
	case bson.D:
		changed := false
		next := make(bson.D, len(v))
		for i, entry := range v {
			childKind := sensitiveField(entry.Key)
			if childKind == "" {
				childKind = kind
			}
			masked, updated, err := m.maskBSON(entry.Value, childKind)
			if err != nil {
				return nil, false, err
			}
			if entry.Key == "_id" && updated {
				return nil, false, fmt.Errorf("the immutable document identifier contains sensitive fields")
			}
			next[i] = bson.E{Key: entry.Key, Value: masked}
			changed = changed || updated
		}
		return next, changed, nil
	case bson.A:
		changed := false
		next := make(bson.A, len(v))
		for i, entry := range v {
			masked, updated, err := m.maskBSON(entry, kind)
			if err != nil {
				return nil, false, err
			}
			next[i], changed = masked, changed || updated
		}
		return next, changed, nil
	case string:
		if kind != "" {
			masked := m.value(kind, v)
			return masked, masked != v, nil
		}
	case int32:
		if kind == piiPhone {
			masked, err := strconv.ParseInt(m.value(kind, strconv.FormatInt(int64(v), 10)), 10, 32)
			return int32(masked), true, err
		}
	case int64:
		if kind == piiPhone {
			masked, err := strconv.ParseInt(m.value(kind, strconv.FormatInt(v, 10)), 10, 64)
			return masked, true, err
		}
	}
	if kind != "" && value != nil {
		return nil, false, fmt.Errorf("unsupported sensitive document field type")
	}
	return value, false, nil
}
