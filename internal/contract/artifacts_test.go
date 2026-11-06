package contract_test

import (
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOpenAPIDocumentsEveryBusinessPath(t *testing.T) {
	content, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		OpenAPI string                    `yaml:"openapi"`
		Paths   map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(content, &document); err != nil {
		t.Fatalf("parse OpenAPI: %v", err)
	}
	if document.OpenAPI == "" {
		t.Fatal("OpenAPI version is missing")
	}
	expected := []string{
		"/auth/login", "/departments", "/departments/{id}", "/employees", "/employees/{id}",
		"/employees/{id}/deactivate", "/leave-requests", "/leave-requests/{id}",
		"/leave-requests/{id}/approve", "/leave-requests/{id}/reject", "/employees/{id}/leave-balances",
		"/attendance/check-in", "/attendance/check-out", "/attendance", "/attendance/report",
	}
	for _, path := range expected {
		if _, ok := document.Paths[path]; !ok {
			t.Errorf("OpenAPI path %s is missing", path)
		}
	}
}

func TestPostmanCollectionIsValidAndComprehensive(t *testing.T) {
	content, err := os.ReadFile("../../postman/HR_Management_API.postman_collection.json")
	if err != nil {
		t.Fatal(err)
	}
	var collection map[string]any
	if err := json.Unmarshal(content, &collection); err != nil {
		t.Fatalf("parse Postman collection: %v", err)
	}
	if count := countRequests(collection["item"]); count < 25 {
		t.Fatalf("Postman collection has %d requests; expected at least 25", count)
	}
}

func countRequests(value any) int {
	items, ok := value.([]any)
	if !ok {
		return 0
	}
	count := 0
	for _, value := range items {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if _, ok := item["request"]; ok {
			count++
		}
		count += countRequests(item["item"])
	}
	return count
}
