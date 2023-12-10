package blueprint

import (
	"capszo-mart/database"
	"capszo-mart/util"
)

type OrderViews int32

const (
	AdminOrderView OrderViews = iota
	CustomerOrderView
	MartOrderView
)

func OrderTransform(order database.Order, view OrderViews) (map[string]interface{}, error) {
	var transformedOrder map[string]interface{}
	var err error

	transformedOrder, err = util.StructToMap(order)
	if err != nil {
		return nil, err
	}

	switch view {
	case AdminOrderView:
		return transformedOrder, nil
	case CustomerOrderView:
		delete(transformedOrder, "customer_id")
		var items []map[string]interface{} = transformedOrder["grocery_items"].([]map[string]interface{})
		for i := range items {
			delete(items[i], "cost_price")
			items[i]["price"] = items[i]["selling_price"]
			delete(items[i], "selling_price")
		}
		transformedOrder["grocery_items"] = items
	case MartOrderView:
		delete(transformedOrder, "customer_id")
		delete(transformedOrder, "mart_id")
		delete(transformedOrder, "otp")
	}

	return transformedOrder, nil
}
