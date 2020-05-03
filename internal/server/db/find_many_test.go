package db

import (
	"github.com/ompluscator/dynamic-struct"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getAge(st interface{}) int64 {
	reader := dynamicstruct.NewReader(st)
	i, _ := reader.GetField("Age").Interface().(int64)
	return i
}

func toAgeList(sts []interface{}) []int64 {
	var ages []int64
	for _, st := range sts {
		ages = append(ages, getAge(st))
	}
	return ages
}

var testData = []string{
	`{"id":"00000000-0000-0000-0000-000000000000",
"type": "user",
"firstName":"Andrew",
"lastName":"Wansley", 
"age": 1}`,
	`{"id":"00000000-0000-0000-0000-000000000000",
"type": "user",
"firstName":"Andrew",
"lastName":"Wansley", 
"age": 2}`,
	`{"id":"00000000-0000-0000-0000-000000000000",
"type": "user",
"firstName":"Andrew",
"lastName":"Wansley", 
"age": 3}`,
}

func TestFindManyApply(t *testing.T) {
	appDB := New()
	AddSampleModels(appDB)

	// add test data
	for _, jsonString := range testData {
		st := makeStruct(appDB, "user", jsonString)
		CreateOperation{Struct: st}.Apply(appDB)
	}
	var findManyTests = []struct {
		operation FindManyOperation
		output    []int64
	}{

		// Simple FindMany
		{
			operation: FindManyOperation{
				ModelName: "user",
				Query: Query{
					FieldCriteria: []FieldCriterion{
						FieldCriterion{
							Key: "Firstname",
							Val: "Andrew",
						},
					},
				},
			},
			output: []int64{1, 2, 3},
		},

		// Simple FindMany
		{
			operation: FindManyOperation{
				ModelName: "user",
				Query: Query{
					FieldCriteria: []FieldCriterion{
						FieldCriterion{
							Key: "Age",
							Val: int64(3),
						},
					},
				},
			},
			output: []int64{3},
		},
	}
	for _, testCase := range findManyTests {
		result := testCase.operation.Apply(appDB)
		rList := result.([]interface{})
		actualAges := toAgeList(rList)
		assert.ElementsMatch(t, testCase.output, actualAges)
	}
}
