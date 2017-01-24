package helpers

import (
	"github.com/pearsonappeng/tensor/db"
	"gopkg.in/mgo.v2/bson"
)

func IsUniqueInventory(name string, OID bson.ObjectId) bool {
	count, err := db.Inventories().Find(bson.M{"name": name, "organization_id": OID}).Count()
	if err == nil && count > 0 {
		return false
	}
	return true
}

func IsNotUniqueInventory(name string, OID bson.ObjectId) bool {
	count, err := db.Inventories().Find(bson.M{"name": name, "organization_id": OID}).Count()
	if err == nil && count > 0 {
		return true
	}
	return false
}

func InventoryExist(ID bson.ObjectId) bool {
	count, err := db.Inventories().FindId(ID).Count()
	if err == nil && count > 0 {
		return true
	}
	return false
}
