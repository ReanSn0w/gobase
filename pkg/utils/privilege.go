package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PrivilegeType int

func (pt *PrivilegeType) Set(privilege PrivilegeType) {
	if pt.Check(privilege) {
		return
	}

	*pt += privilege
}

func (pt *PrivilegeType) Unset(privilege PrivilegeType) {
	if !pt.Check(privilege) {
		return
	}

	*pt -= privilege
}

func (pt *PrivilegeType) Check(privilege PrivilegeType) bool {
	return int(*pt)&int(privilege) == 1
}

const (
	OwnerRead    PrivilegeType = 1   // Разрешение на чтение
	OwnerWrite   PrivilegeType = 2   // Разрешение на создание новых документов
	OwnerUpdate  PrivilegeType = 4   // Разрешение на обновление существующих документов
	OwnerDelete  PrivilegeType = 8   // Разрешение на удаление документов
	PublicRead   PrivilegeType = 16  // Разрешение на чтение чужих материалов
	PublicWrite  PrivilegeType = 32  // Разрешение на создание материалов от имени другиз пользователей
	PublicUpdate PrivilegeType = 64  // Разрешение на обновлнеие чужих материалов
	PublicDelete PrivilegeType = 128 // Разрешение на удаление чужих материалов
)

var (
	privileges = newPrivilegesStorage()
)

// Получает структуру для работы с привилегиями
func Privileges() *PrivilegesStorage {
	return privileges
}

func newPrivilegesStorage() *PrivilegesStorage {
	objID, _ := db.PredictableObjectID("privileges")
	storage := &PrivilegesStorage{ID: objID}

	storage.SetGroup("guest", "main", PublicRead)
	storage.SetGroup("banned", "main", PublicRead, OwnerRead)
	storage.SetGroup("user", "main", OwnerRead, OwnerWrite, OwnerUpdate, OwnerDelete, PublicRead)
	storage.SetGroup("moderator", "main", OwnerRead, OwnerWrite, OwnerUpdate, OwnerDelete, PublicRead, PublicUpdate)
	storage.SetGroup("admin", "main", OwnerRead, OwnerWrite, OwnerUpdate, OwnerDelete, PublicRead, PublicWrite, PublicUpdate, OwnerDelete)

	return storage
}

type PrivilegesStorage struct {
	ID        primitive.ObjectID       `bson:"_id"`
	Rules     map[string]PrivilegeType `bson:"rules"`
	Timestamp time.Time                `bson:"time"`

	shared bool
}

func (ps *PrivilegesStorage) Sync() {
	ps.shared = true

	go ps.autoupdate()
}

func (ps *PrivilegesStorage) Unsync() {
	ps.shared = false
}

func (ps *PrivilegesStorage) autoupdate() {
	for ps.shared {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)

		c, _ := DB().Collection(systemCollection).CountDocuments(ctx, bson.D{
			{Key: "_id", Value: ps.ID},
			{Key: "time", Value: ps.Timestamp},
		})

		if c > 0 {
			cancel()
		} else {
			ps.Timestamp = time.Now()

			_, err := DB().Collection(systemCollection).UpdateByID(ctx, ps.ID, bson.D{
				{Key: "rules", Value: ps.Rules},
				{Key: "time", Value: ps.Timestamp},
			})

			if err != nil {
				log.Println(err)
			}

			cancel()
		}
	}
}

func (ps *PrivilegesStorage) SetGroup(name, module string, privileges ...PrivilegeType) {
	ps.updatemask(ps.generateKey(name, module), func(pt PrivilegeType) PrivilegeType {
		for _, item := range privileges {
			pt.Set(item)
		}

		return pt
	})
}

func (ps *PrivilegesStorage) UnsetGroup(name, module string, privileges ...PrivilegeType) {
	ps.updatemask(ps.generateKey(name, module), func(pt PrivilegeType) PrivilegeType {
		for _, item := range privileges {
			pt.Unset(item)
		}

		return pt
	})
}

func (ps *PrivilegesStorage) SetUser(id primitive.ObjectID, module string, privileges ...PrivilegeType) {
	ps.updatemask(ps.generateKey(id.Hex(), module), func(pt PrivilegeType) PrivilegeType {
		for _, item := range privileges {
			pt.Set(item)
		}

		return pt
	})
}

func (ps *PrivilegesStorage) UnsetUser(id primitive.ObjectID, module string, privileges ...PrivilegeType) {
	ps.updatemask(ps.generateKey(id.Hex(), module), func(pt PrivilegeType) PrivilegeType {
		for _, item := range privileges {
			pt.Unset(item)
		}

		return pt
	})
}

func (ps *PrivilegesStorage) Check(id primitive.ObjectID, group, module string, privileges ...PrivilegeType) bool {
	val, err := ps.check(id.Hex(), module, privileges...)
	if err == nil {
		return val
	}

	val, err = ps.check(id.Hex(), "main", privileges...)
	if err == nil {
		return val
	}

	val, err = ps.check(group, module, privileges...)
	if err == nil {
		return val
	}

	val, err = ps.check(group, "main", privileges...)
	if err == nil {
		return val
	}

	return false
}

func (ps *PrivilegesStorage) check(first, second string, privileges ...PrivilegeType) (bool, error) {
	value, err := ps.getmask(ps.generateKey(first, second))
	if err != nil {
		return false, err
	}

	for _, item := range privileges {
		if !value.Check(item) {
			return false, nil
		}
	}

	return true, nil
}

func (ps *PrivilegesStorage) getmask(key string) (PrivilegeType, error) {
	val, b := ps.Rules[key]
	if !b {
		return val, errors.New("value not found")
	}

	return val, nil
}

func (ps *PrivilegesStorage) setmask(key string, value PrivilegeType) {
	ps.Rules[key] = value
	ps.save()
}

func (ps *PrivilegesStorage) updatemask(key string, update func(PrivilegeType) PrivilegeType) {
	mask, _ := ps.getmask(key)
	mask = update(mask)
	ps.setmask(key, mask)
}

func (ps *PrivilegesStorage) generateKey(first, second string) string {
	return fmt.Sprintf("%s.%s", first, second)
}

func (ps *PrivilegesStorage) save() error {
	if !ps.shared {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	_, err := DB().Collection(systemCollection).UpdateByID(ctx, ps.ID, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "rules", Value: ps.Rules},
		}},
	})

	return err
}
