package godiff

import (
	"fmt"
	"reflect"
)

type DiffResult struct {
	Field string
	AVal  interface{}
	BVal  interface{}
}

func argToValue(a interface{}) (*reflect.Value, error) {
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, fmt.Errorf("arg is nil")
		} else {
			v = reflect.Indirect(v)
		}
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("isn't a struct")
	}

	return &v, nil
}

func StructDiff(a, b interface{}) ([]*DiffResult, error) {
	av, err := argToValue(a)
	if err != nil {
		return nil, err
	}

	bv, err := argToValue(b)
	if err != nil {
		return nil, err
	}

	results := make([]*DiffResult, 0)

	err = structDiff("", av, bv, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func structDiff(prefix string, av, bv *reflect.Value, results *[]*DiffResult) error {
	t := av.Type() // doesn't matter if we pick a or b for this
	nf := av.NumField()
	for i := 0; i < nf; i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue // ignore unexported fields
		}

		afv := av.Field(i)
		bfv := bv.Field(i)
		fname := field.Name

		if !afv.CanInterface() {
			return fmt.Errorf("field %s: !CanInterface()", fname)
		}

		afvi := afv.Interface()
		bfvi := bfv.Interface()

		if reflect.DeepEqual(afvi, bfvi) {
			continue
		}

		if field.Anonymous {
			err := structDiff(fname+".", &afv, &bfv, results)
			if err != nil {
				return err
			}
			continue
		}

		result := &DiffResult{
			Field: prefix + fname,
			AVal:  afvi,
			BVal:  bfvi,
		}

		*results = append(*results, result)
	}

	return nil
}

func MapDiff(a, b interface{}) ([]*DiffResult, error) {
	av := reflect.ValueOf(a)
	if av.Kind() != reflect.Map {
		return nil, fmt.Errorf("a is not a map")
	}

	bv := reflect.ValueOf(b)
	if bv.Kind() != reflect.Map {
		return nil, fmt.Errorf("b is not a map")
	}

	results := make([]*DiffResult, 0)

	err := mapDiff(av, bv, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func mapDiff(av, bv reflect.Value, results *[]*DiffResult) error {
	iter := av.MapRange()
	for iter.Next() {
		amk := iter.Key()
		amv := iter.Value()
		if !amv.CanInterface() {
			return fmt.Errorf("A %s: !CanInterface()", amk.String())
		}
		amvi := amv.Interface()

		bmv := bv.MapIndex(amk)
		if !bmv.IsValid() {
			// key exists in A but not B
			result := &DiffResult{
				Field: amk.String(),
				AVal:  amv.Interface(),
				BVal:  nil,
			}
			*results = append(*results, result)
			continue
		}

		if !bmv.CanInterface() {
			return fmt.Errorf("B %s: !CanInterface()", amk.String())
		}
		bmvi := bmv.Interface()

		if !reflect.DeepEqual(amvi, bmvi) {
			// key exists in A and B, but values differ
			result := &DiffResult{
				Field: amk.String(),
				AVal:  amvi,
				BVal:  bmvi,
			}
			*results = append(*results, result)
		}
	}

	iter = bv.MapRange()
	for iter.Next() {
		bmk := iter.Key()
		bmv := iter.Value()
		if !bmv.CanInterface() {
			return fmt.Errorf("BB %s: !CanInterface()", bmk.String())
		}

		amv := av.MapIndex(bmk)
		if !amv.IsValid() {
			// key exists in B but not A
			result := &DiffResult{
				Field: bmk.String(),
				AVal:  nil,
				BVal:  bmv.Interface(),
			}
			*results = append(*results, result)
		}
	}

	return nil
}
