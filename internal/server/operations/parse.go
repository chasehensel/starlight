package operations

import (
	"awans.org/aft/internal/db"
	"awans.org/aft/internal/model"
	"errors"
	"fmt"
)

var (
	ErrParse            = errors.New("parse-error")
	ErrUnusedKeys       = fmt.Errorf("%w: unused keys", ErrParse)
	ErrInvalidModel     = fmt.Errorf("%w: invalid model", ErrParse)
	ErrInvalidStructure = fmt.Errorf("%w: invalid-structure", ErrParse)
)

type void struct{}
type set map[string]void

func (s set) String() string {
	var ss []string
	for k := range s {
		ss = append(ss, k)
	}
	return fmt.Sprintf("%v", ss)
}

type Parser struct {
	tx db.Tx
}

// parseAttribute tries to consume an attribute key from a json map; returns whether the attribute was consumed
func parseAttribute(key string, data map[string]interface{}, rec model.Record) bool {
	value, ok := data[key]
	if ok {
		rec.Set(key, value)
	}
	return ok
}

func (p Parser) parseNestedCreate(r model.Relationship, data map[string]interface{}) (op NestedOperation, err error) {
	unusedKeys := make(set)
	for k := range data {
		unusedKeys[k] = void{}
	}

	m, err := p.tx.GetModel(r.TargetModel)
	if err != nil {
		return
	}
	rec, unusedKeys := buildRecordFromData(m, unusedKeys, data)
	nested := []NestedOperation{}
	for k, r := range m.Relationships {
		additionalNested, consumed, err := p.parseRelationship(k, r, data)
		if err != nil {
			return NestedCreateOperation{}, err
		}
		if consumed {
			delete(unusedKeys, k)
		}
		nested = append(nested, additionalNested...)
	}
	if len(unusedKeys) != 0 {
		return NestedCreateOperation{}, fmt.Errorf("%w: %v", ErrUnusedKeys, unusedKeys)
	}
	nestedCreate := NestedCreateOperation{Relationship: r, Record: rec, Nested: nested}
	return nestedCreate, nil
}

func parseNestedConnect(r model.Relationship, data map[string]interface{}) NestedConnectOperation {
	if len(data) != 1 {
		panic("Too many keys in a unique query")
	}
	// this should be a separate method
	var uq UniqueQuery
	for k, v := range data {
		sv := v.(string)
		uq = UniqueQuery{Key: k, Val: sv}
	}
	return NestedConnectOperation{Relationship: r, UniqueQuery: uq}
}

func listify(val interface{}) []interface{} {
	var opList []interface{}
	switch v := val.(type) {
	case map[string]interface{}:
		opList = []interface{}{v}
	case []interface{}:
		opList = v
	default:
		panic("Invalid input")
	}
	return opList
}

func (p Parser) parseRelationship(key string, r model.Relationship, data map[string]interface{}) ([]NestedOperation, bool, error) {
	nestedOpMap, ok := data[key].(map[string]interface{})
	if !ok {
		_, isValue := data[key]
		if !isValue {
			return []NestedOperation{}, false, nil
		}

		return []NestedOperation{}, false, fmt.Errorf("%w: expected an object, got: %v", ErrInvalidStructure, data)
	}
	var nested []NestedOperation
	for k, val := range nestedOpMap {
		opList := listify(val)
		for _, op := range opList {
			nestedOp, ok := op.(map[string]interface{})
			if !ok {
				return nil, false, fmt.Errorf("%w: expected an object, got: %v", ErrInvalidStructure, nestedOp)
			}
			switch k {
			case "connect":
				nestedConnect := parseNestedConnect(r, nestedOp)
				nested = append(nested, nestedConnect)
			case "create":
				nestedCreate, err := p.parseNestedCreate(r, nestedOp)
				if err != nil {
					return nil, false, err
				}
				nested = append(nested, nestedCreate)
			}
		}
	}

	return nested, true, nil
}

func buildRecordFromData(m model.Model, keys set, data map[string]interface{}) (model.Record, set) {
	rec := model.RecordForModel(m)
	for k := range m.Attributes {
		if parseAttribute(k, data, rec) {
			delete(keys, k)
		}
	}
	return rec, keys
}

func (p Parser) ParseCreate(modelName string, data map[string]interface{}) (op CreateOperation, err error) {
	unusedKeys := make(set)
	for k := range data {
		unusedKeys[k] = void{}
	}

	m, err := p.tx.GetModel(modelName)
	if err != nil {
		return op, fmt.Errorf("%w: %v", ErrInvalidModel, modelName)
	}
	rec, unusedKeys := buildRecordFromData(m, unusedKeys, data)
	nested := []NestedOperation{}
	for k, r := range m.Relationships {
		additionalNested, consumed, err := p.parseRelationship(k, r, data)
		if err != nil {
			return op, err
		}
		if consumed {
			delete(unusedKeys, k)
		}
		nested = append(nested, additionalNested...)
	}
	if len(unusedKeys) != 0 {
		return op, fmt.Errorf("%w: %v", ErrUnusedKeys, unusedKeys)
	}
	op = CreateOperation{Record: rec, Nested: nested}
	return op, err
}

func (p Parser) ParseFindOne(modelName string, data map[string]interface{}) (op FindOneOperation, err error) {
	m, err := p.tx.GetModel(modelName)
	if err != nil {
		return
	}
	var fieldName string
	var value interface{}

	if len(data) > 1 {
		panic("too much data in findOne")
	} else if len(data) == 0 {
		panic("empty data in findOne")
	}

	for k, v := range data {
		attr := m.GetAttributeByJsonName(k)
		fieldName = model.JsonKeyToFieldName(k)
		value = attr.ParseFromJson(v)
	}

	op = FindOneOperation{
		UniqueQuery: UniqueQuery{
			Key: fieldName,
			Val: value,
		},
		ModelName: modelName,
	}
	return op, nil
}

func (p Parser) ParseFindMany(modelName string, data map[string]interface{}) (op FindManyOperation, err error) {
	q, err := p.ParseQuery(modelName, data)
	if err != nil {
		return
	}

	op = FindManyOperation{
		Query:     q,
		ModelName: modelName,
	}
	return op, nil
}

func (p Parser) parseCompositeQueryList(modelName string, opVal interface{}) (ql []Query, err error) {
	opList := opVal.([]interface{})
	for _, opData := range opList {
		opMap := opData.(map[string]interface{})
		var opQ Query
		opQ, err = p.ParseQuery(modelName, opMap)
		if err != nil {
			return
		}
		ql = append(ql, opQ)
	}
	return
}

func (p Parser) ParseQuery(modelName string, data map[string]interface{}) (q Query, err error) {
	m, err := p.tx.GetModel(modelName)
	if err != nil {
		return
	}
	q = Query{}
	fc := parseFieldCriteria(m, data)
	q.FieldCriteria = fc
	rc, err := p.parseSingleRelationshipCriteria(m, data)
	if err != nil {
		return
	}
	q.RelationshipCriteria = rc
	arc, err := p.parseAggregateRelationshipCriteria(m, data)
	if err != nil {
		return
	}
	q.AggregateRelationshipCriteria = arc

	if orVal, ok := data["OR"]; ok {
		var orQL []Query
		orQL, err = p.parseCompositeQueryList(modelName, orVal)
		if err != nil {
			return
		}
		q.Or = orQL
	}
	if andVal, ok := data["AND"]; ok {
		var andQL []Query
		andQL, err = p.parseCompositeQueryList(modelName, andVal)
		if err != nil {
			return
		}
		q.And = andQL
	}
	if notVal, ok := data["NOT"]; ok {
		var notQL []Query
		notQL, err = p.parseCompositeQueryList(modelName, notVal)
		if err != nil {
			return
		}
		q.Not = notQL
	}
	return
}

func (p Parser) parseSingleRelationshipCriteria(m model.Model, data map[string]interface{}) (rcl []RelationshipCriterion, err error) {
	for k, rel := range m.Relationships {
		if rel.RelType == model.HasOne || rel.RelType == model.BelongsTo {
			if value, ok := data[k]; ok {
				var rc RelationshipCriterion
				rc, err = p.parseRelationshipCriterion(rel, value)
				if err != nil {
					return
				}
				rcl = append(rcl, rc)
			}
		}
	}
	return rcl, nil
}

func (p Parser) parseAggregateRelationshipCriteria(m model.Model, data map[string]interface{}) (arcl []AggregateRelationshipCriterion, err error) {
	for k, rel := range m.Relationships {
		if rel.RelType == model.HasMany || rel.RelType == model.HasManyAndBelongsToMany {
			if value, ok := data[k]; ok {
				var arc AggregateRelationshipCriterion
				arc, err = p.parseAggregateRelationshipCriterion(rel, value)
				if err != nil {
					return
				}
				arcl = append(arcl, arc)
			}
		}
	}
	return arcl, nil
}

func parseFieldCriteria(m model.Model, data map[string]interface{}) []FieldCriterion {
	var fieldCriteria []FieldCriterion
	for k, attr := range m.Attributes {
		if value, ok := data[k]; ok {
			fc := parseFieldCriterion(k, attr, value)
			fieldCriteria = append(fieldCriteria, fc)
		}
	}
	return fieldCriteria
}

func parseFieldCriterion(key string, a model.Attribute, value interface{}) FieldCriterion {
	fieldName := model.JsonKeyToFieldName(key)
	parsedValue := a.ParseFromJson(value)
	fc := FieldCriterion{
		// TODO handle function values like {startsWith}
		Key: fieldName,
		Val: parsedValue,
	}
	return fc
}

func (p Parser) parseAggregateRelationshipCriterion(r model.Relationship, value interface{}) (arc AggregateRelationshipCriterion, err error) {
	mapValue := value.(map[string]interface{})
	if len(mapValue) > 1 {
		panic("too much data in parseAggregateRel")
	} else if len(mapValue) == 0 {
		panic("empty data in parseAggregateRel")
	}
	var ag Aggregation
	for k, v := range mapValue {
		switch k {
		case "some":
			ag = Some
		case "none":
			ag = None
		case "every":
			ag = Every
		default:
			panic("Bad aggregation")
		}
		var rc RelationshipCriterion
		rc, err = p.parseRelationshipCriterion(r, v)
		if err != nil {
			return
		}
		arc = AggregateRelationshipCriterion{
			Aggregation:           ag,
			RelationshipCriterion: rc,
		}
	}
	return
}

func (p Parser) parseRelationshipCriterion(r model.Relationship, value interface{}) (rc RelationshipCriterion, err error) {
	mapValue := value.(map[string]interface{})
	m, err := p.tx.GetModel(r.TargetModel)
	if err != nil {
		return
	}
	fc := parseFieldCriteria(m, mapValue)
	rrc, err := p.parseSingleRelationshipCriteria(m, mapValue)
	if err != nil {
		return
	}
	arrc, err := p.parseAggregateRelationshipCriteria(m, mapValue)
	if err != nil {
		return
	}
	rc = RelationshipCriterion{
		Relationship:                         r,
		RelatedFieldCriteria:                 fc,
		RelatedRelationshipCriteria:          rrc,
		RelatedAggregateRelationshipCriteria: arrc,
	}
	return
}

func (p Parser) parseInclusion(r model.Relationship, value interface{}) Inclusion {
	if v, ok := value.(bool); ok {
		if v {
			return Inclusion{Relationship: r, Query: Query{}}
		} else {
			panic("Include specified as false?")
		}
	}
	panic("Include with findMany args not yet implemented")
}

func (p Parser) ParseInclude(modelName string, data map[string]interface{}) (i Include, err error) {
	m, err := p.tx.GetModel(modelName)
	if err != nil {
		return
	}
	var includes []Inclusion
	for k, val := range data {
		rel := m.Relationships[k]
		inc := p.parseInclusion(rel, val)
		includes = append(includes, inc)
	}
	i = Include{Includes: includes}
	return
}
