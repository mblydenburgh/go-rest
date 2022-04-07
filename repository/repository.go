package repository

import (
	"log"
	"mblydenburgh/go-rest/domain"

	"github.com/guregu/dynamo"
	uuid "github.com/satori/go.uuid"
)

var tableName = "go-rest-table"

func GetCar(client *dynamo.DB, id string) (*CarItem, error) {
	log.Printf("Looking up car id %v on table %v", id, tableName)
	table := client.Table(tableName)

	var result CarItem
	rangeKey := "Car#"
	err := table.Get("UserId", id).Range("ModelTypeAndId", dynamo.BeginsWith, rangeKey).One(&result)
	if err != nil {
		log.Println("error getting car")
		log.Println(err)
		return nil, err
	}

	log.Println("found car")
	return &result, nil
}

func PutCar(client *dynamo.DB, car *domain.SaveCarPayload) (string, error) {
	log.Println("Saving car")
	table := client.Table(tableName)
	uuid := uuid.NewV4().String()
	modelTypeAndId := "Car#" + uuid
	item := CarItem{
		UserId:         uuid,
		ModelTypeAndId: modelTypeAndId,
		Manufacturer:   car.Manufacturer,
		Model:          car.Model,
		Year:           car.Year,
		Trim:           car.Trim,
		VehicleType:    car.VehicleType,
		Color:          car.Color,
		VIN:            car.VIN,
	}
	putAction := table.Put(item)
	err := putAction.Run()
	if err != nil {
		log.Printf("Error getting item: %v", err)
		return "", err
	}
	return uuid, nil
}

func DeleteCar(client *dynamo.DB, id string) error {
	log.Printf("Deleting car %v", id)
	table := client.Table(tableName)
	rangeKey := "Car#" + id
	deleteAction := table.Delete("UserId", id).Range("ModelTypeAndId", rangeKey)
	err := deleteAction.Run()
	if err != nil {
		log.Printf("Error deleting car %v, ", err)
		return err
	}

	log.Println("Deleted car")
	return nil
}
