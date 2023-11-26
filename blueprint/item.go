package blueprint

import (
	"capszo-mart/database"
	"capszo-mart/util"
)

type itemViews int32

const (
	AdminItem itemViews = iota
	CustomerItem
	MartItem
)

func ItemTransform(item database.Item, view itemViews) (map[string]interface{}, error) {
	var transformedItem map[string]interface{}
	var err error

	transformedItem, err = util.StructToMap(item)
	if err != nil {
		return nil, err
	}

	switch view {
	case AdminItem:
		return transformedItem, nil
	case CustomerItem:
		delete(transformedItem, "cost_price")
		delete(transformedItem, "stock_quantity")
		transformedItem["price"] = transformedItem["selling_price"]
		delete(transformedItem, "selling_price")
		delete(transformedItem, "created_at")
		delete(transformedItem, "updated_at")
	case MartItem:
		delete(transformedItem, "created_at")
		delete(transformedItem, "updated_at")
	}

	return transformedItem, nil
}
