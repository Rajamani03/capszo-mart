package blueprint

import (
	"capszo-mart/database"
	"capszo-mart/util"
)

type orderViews int32

const (
	AdminOrder orderViews = iota
	CustomerOrder
	MartOrder
)

func OrderTransform(order database.Order, view orderViews) (map[string]interface{}, error) {
	var transformedOrder map[string]interface{}
	var err error

	transformedOrder, err = util.StructToMap(order)
	if err != nil {
		return nil, err
	}

	switch view {
	case AdminOrder:
		return transformedOrder, nil
	case CustomerOrder:
		delete(transformedOrder, "customer_id")
	case MartOrder:
		delete(transformedOrder, "customer_id")
		delete(transformedOrder, "mart_id")
		delete(transformedOrder, "otp")
	}

	return transformedOrder, nil
}
