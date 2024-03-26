package godiff

import (
	"testing"
)

type ProxyInstanceConfig struct {
	VMType   string
	Metadata map[string]string
}

type ProxyInstance struct {
	ProxyInstanceConfig
	CfgState string
}

func checkNoDiffs(t *testing.T, a, b *ProxyInstance) {
	results, err := StructDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if results == nil {
		t.Fatalf("expected an empty array got nil")
	}

	if len(results) != 0 {
		t.Fatalf("expected zero diffs, got %d", len(results))
	}
}

func TestNoDiff(t *testing.T) {
	a := &ProxyInstance{}
	b := &ProxyInstance{}

	checkNoDiffs(t, a, b)
}

func TestNil(t *testing.T) {
	var a *ProxyInstance
	var b *ProxyInstance

	_, err := StructDiff(a, b)
	if err == nil {
		t.Fatalf("expected err")
	}

	b = &ProxyInstance{}
	_, err = StructDiff(a, b)
	if err == nil {
		t.Fatalf("expected err")
	}

	a = &ProxyInstance{}
	b = nil
	_, err = StructDiff(a, b)
	if err == nil {
		t.Fatalf("expected err")
	}
}

func TestStringDiff(t *testing.T) {
	a := &ProxyInstance{
		CfgState: "starting",
	}
	b := &ProxyInstance{
		CfgState: "running",
	}

	results, err := StructDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected one diff, got %d", len(results))
	}

	result := results[0]
	if result.Field != "CfgState" {
		t.Fatalf("expected Field == 'CfgState', got %#v", *result)
	}

	if result.AVal.(string) != "starting" || result.BVal.(string) != "running" {
		t.Fatalf("value mismatch, got %#v", *result)
	}
}

func TestMultiStringDiff(t *testing.T) {
	a := &ProxyInstance{
		ProxyInstanceConfig: ProxyInstanceConfig{
			VMType: "vmtype-a",
		},
		CfgState: "starting",
	}
	b := &ProxyInstance{
		ProxyInstanceConfig: ProxyInstanceConfig{
			VMType: "vmtype-b",
		},
		CfgState: "running",
	}

	results, err := StructDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(results))
	}

	result := results[0]
	if result.Field != "ProxyInstanceConfig.VMType" {
		t.Fatalf("expected Field == 'VMType', got %#v", *result)
	}

	if result.AVal != "vmtype-a" || result.BVal != "vmtype-b" {
		t.Fatalf("value mismatch, got %#v", *result)
	}

	result = results[1]
	if result.Field != "CfgState" {
		t.Fatalf("expected Field == 'CfgState', got %#v", *result)
	}

	if result.AVal.(string)!= "starting" || result.BVal.(string) != "running" {
		t.Fatalf("value mismatch, got %#v", *result)
	}

}

func TestMetadataStructDiff(t *testing.T) {
	a := &ProxyInstance{
		ProxyInstanceConfig: ProxyInstanceConfig{Metadata: map[string]string{"key1": "arg1", "key2": "arg2"}},
	}
	b := &ProxyInstance{
		ProxyInstanceConfig: ProxyInstanceConfig{Metadata: map[string]string{"key1": "arg1", "key2": "arg2"}},
	}

	results, err := StructDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 diffs, got %d", len(results))
	}

	a = &ProxyInstance{
		ProxyInstanceConfig: ProxyInstanceConfig{Metadata: map[string]string{"key1": "arg1", "key2": "arg2", "key3": "arg3"}},
	}

	results, err = StructDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 diffs, got %d", len(results))
	}

	result := results[0]
	if result.Field != "ProxyInstanceConfig.Metadata" {
		t.Logf("result: %#v", *result)
		t.Fatalf("bad field")
	}

	_, aok := result.AVal.(map[string]string)
	if !aok {
		t.Logf("result: %#v", *result)
		t.Fatalf("aval bad type")
	}

	_, bok := result.AVal.(map[string]string)
	if !bok {
		t.Logf("result: %#v", *result)
		t.Fatalf("bval bad type")
	}

	results, err = MapDiff(result.AVal, result.BVal)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 map diff, got %d", len(results))
	}

	result = results[0]
	if result.Field != "key3" {
		t.Fatalf("expected field to be key3")
	}
	if result.AVal != "arg3" {
		t.Fatalf("expected aval == 'arg3' got %v", result.AVal)
	}

	if result.BVal != nil {
		t.Fatalf("expected bval == nil got %v", result.BVal)
	}
}

func TestMapDiff(t *testing.T) {
	a := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}
	b := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}

	// a == b
	results, err := MapDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 diffs, got %d", len(results))
	}

	delete(b, "key3")

	// key in a but not in b
	results, err = MapDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 map diff, got %d", len(results))
	}

	result := results[0]
	if result.Field != "key3" {
		t.Fatalf("expected field to be key3")
	}
	if result.AVal != "val3" {
		t.Fatalf("expected aval == 'arg3' got %v", result.AVal)
	}

	if result.BVal != nil {
		t.Fatalf("expected bval == nil got %v", result.BVal)
	}

	a = map[string]string{
		"key1": "val1",
	}
	b = map[string]string{
		"key1": "val1",
		"key2": "val2",
	}

	// key in b but not in a
	results, err = MapDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 map diff, got %d", len(results))
	}

	result = results[0]
	if result.Field != "key2" {
		t.Fatalf("expected field to be key2")
	}
	if result.AVal != nil {
		t.Fatalf("expected aval == nil got %v", result.AVal)
	}
	if result.BVal != "val2" {
		t.Fatalf("expected bval == 'val2' got %v", result.BVal)
	}

	// key in both but values differ
	a = map[string]string{
		"key1": "a",
	}
	b = map[string]string{
		"key1": "b",
	}

	results, err = MapDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 map diff, got %d", len(results))
	}

	result = results[0]
	if result.Field != "key1" {
		t.Fatalf("expected field to be key1, got %v", result.Field)
	}
	if result.AVal != "a" {
		t.Fatalf("expected aval == a got %v", result.AVal)
	}
	if result.BVal != "b" {
		t.Fatalf("expected bval == 'b' got %v", result.BVal)
	}
}

func TestArrayDiff(t *testing.T) {
	a := []string{"a", "b", "c"}

	results, err := ArrayDiff(a, a)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no diff, found %d", len(results))
	}

	b := []string{"a", "b", "b"}

	results, err = ArrayDiff(a, b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 diff, found %d", len(results))
	}
	t.Logf("%#v", results[0])
}
