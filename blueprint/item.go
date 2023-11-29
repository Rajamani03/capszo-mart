package blueprint

import (
	"capszo-mart/database"
	"capszo-mart/util"
)

type ItemViews int32

const (
	AdminItemView ItemViews = iota
	CustomerItemView
	MartItemView
)

func ItemTransform(item database.Item, view ItemViews) (map[string]interface{}, error) {
	var transformedItem map[string]interface{}
	var err error

	transformedItem, err = util.StructToMap(item)
	if err != nil {
		return nil, err
	}

	switch view {
	case AdminItemView:
		return transformedItem, nil
	case CustomerItemView:
		delete(transformedItem, "cost_price")
		delete(transformedItem, "stock_quantity")
		transformedItem["price"] = transformedItem["selling_price"]
		delete(transformedItem, "selling_price")
		delete(transformedItem, "created_at")
		delete(transformedItem, "updated_at")
	case MartItemView:
		delete(transformedItem, "created_at")
		delete(transformedItem, "updated_at")
	}

	return transformedItem, nil
}
