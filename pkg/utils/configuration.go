package utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ReanSn0w/mongo-monkey/wrap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Configuration = &config{}
)

type config struct{}

// Заирузка конфигурации из доцумента с конфигами
func (config *config) Load(module string, object interface{}) error {
	return DB().Operation(func(ctx context.Context, w *wrap.Wrap) error {
		c := w.Collection(systemCollection)

		cur, err := c.Aggregate(ctx, mongo.Pipeline{
			{{Key: "$match", Value: bson.D{{Key: "_id", Value: config.objectID(w)}}}},
			{{Key: "$project", Value: config.projectStep(module, object)}},
		})
		if err != nil {
			return err
		}

		cur.Next(ctx)
		return cur.Decode(object)
	})
}

// Сохранение настроек модуля
func (config *config) Save(module string, object interface{}) error {
	return DB().UpdateObj(config.objectID(DB()), systemCollection, bson.D{
		{Key: "$set", Value: bson.D{{Key: module, Value: object}}},
	})
}

func (config *config) objectID(w *wrap.Wrap) primitive.ObjectID {
	oid, _ := w.PredictableObjectID("configuration")
	return oid
}

func (config *config) projectStep(module string, object interface{}) bson.D {
	ro := reflect.TypeOf(object)
	fieldsCount := ro.NumField()
	project := []bson.E{}

	// Получение названия полей объекта
	for i := 0; i < fieldsCount; i++ {
		field := ro.Field(i)
		tag := field.Tag

		val, ok := tag.Lookup("bson")
		if !ok {
			val = field.Name
		}

		project = append(project, bson.E{Key: val, Value: fmt.Sprintf("%s.%s", module, val)})
	}

	return bson.D(project)
}
