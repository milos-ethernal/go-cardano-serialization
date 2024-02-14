package components_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	myMap := map[interface{}]interface{}{
		0: map[interface{}]interface{}{
			1: map[string]interface{}{
				"chainId": "vector",
				"transactions": []map[string]interface{}{
					{"address": "addr_test1vzqcvg32jjgvq6sc52l55fdfgrl369w9wyglqfysj4ey3cqs45ck7", "amount": 1000000},
					{"address": "addr_test1vrtrch74dk7u5m5gjl76p5yfrlascgmaqplzpfe6xn9pxkqsf0l04", "amount": 200000},
				},
			},
		},
	}

	uintMap := convertToUintMap(myMap)

	//fmt.Println(uintMap[0])

	// Assert uintMap[0] to map[uint]interface{}
	innerMap, ok := uintMap[0].(map[uint]interface{})
	if !ok {
		fmt.Println("uintMap[0] is not of type map[uint]interface{}")
		return
	}

	//fmt.Println("Inner map:", innerMap[1])

	innerInnerMap, ok := innerMap[1].(map[string]interface{})
	if !ok {
		fmt.Println("innerMap[1] is not of type map[string]interface{}")
		return
	}

	fmt.Println("chainId:", innerInnerMap["chainId"])

	transactions, ok := innerInnerMap["transactions"].([]map[string]interface{})
	if !ok {
		fmt.Println("innerInnerMap[transactions] is not of type []map[string]interface{}")
		return
	}

	for _, transaction := range transactions {
		for key, value := range transaction {
			fmt.Print(key + " - ")
			fmt.Println(value)
		}
	}

	assert.Equal(t, 1, 2)
}

func convertToUintMap(inputMap map[interface{}]interface{}) map[uint]interface{} {
	outputMap := make(map[uint]interface{})

	for k, v := range inputMap {
		// Convert key to uint if possible
		var newKey uint
		switch k := k.(type) {
		case uint:
			newKey = k
		case int:
			newKey = uint(k)
		default:
			// Handle unsupported key type
			panic(fmt.Sprintf("Unsupported key type: %v", reflect.TypeOf(k)))
		}

		// Convert value recursively if it's a map
		if nestedMap, ok := v.(map[interface{}]interface{}); ok {
			outputMap[newKey] = convertToUintMap(nestedMap)
		} else {
			outputMap[newKey] = v
		}
	}

	return outputMap
}
