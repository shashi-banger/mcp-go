package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ListToolsPP(response *json.RawMessage) (*json.RawMessage, error) {
	// Transform the response of list tools such that the inputSchema
	// does not have any references.

	responseMap := make(map[string]interface{})
	err := json.Unmarshal(*response, &responseMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	tools := responseMap["tools"].([]interface{})

	for _, tool := range tools {
		toolMap := tool.(map[string]interface{})
		inputSchema := toolMap["inputSchema"].(map[string]interface{})
		unrolledInputSchema, err := UnrollRefs(inputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to unroll input schema: %w", err)
		}

		toolMap["inputSchema"] = unrolledInputSchema
	}

	delete(responseMap, "$defs")
	// Marshal the response map back to a JSON string
	responseJSON, err := json.Marshal(responseMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return (*json.RawMessage)(&responseJSON), nil
}

// UnrollRefs takes a JSON schema represented as a map and replaces all $ref
// pointers with their corresponding definitions found within the $defs section.
// It returns the modified schema with $refs unrolled and the $defs section removed.
// This function is generic and works on any valid JSON structure represented as
// map[string]interface{} or []interface{}.
func UnrollRefs(schema map[string]interface{}) (map[string]interface{}, error) {
	// Extract the $defs section if it exists
	defs, ok := schema["$defs"].(map[string]interface{})
	if !ok {
		// No $defs section, return the schema as is.
		// If there were external refs, they would remain unresolved here.
		return schema, nil
	}

	// Create a copy of the schema excluding $defs to start the unrolling process
	mainSchema := make(map[string]interface{})
	for key, value := range schema {
		if key != "$defs" {
			mainSchema[key] = value
		}
	}

	// Recursively unroll references starting from the main schema body
	unrolledSchema := unrollRecursive(mainSchema, defs)

	// The recursive function returns the unrolled structure.
	// We expect the top level to still be a map.
	result, ok := unrolledSchema.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unrolled schema is not a map")
	}

	return result, nil
}

// unrollRecursive is a helper function that recursively traverses the schema
// and replaces $ref pointers using the provided definitions.
func unrollRecursive(schema interface{}, defs map[string]interface{}) interface{} {
	switch v := schema.(type) {
	case map[string]interface{}:
		// Check if the current map contains a $ref
		if ref, ok := v["$ref"].(string); ok {
			// Resolve the reference string to find the actual definition
			resolvedDef := resolveRef(ref, defs)
			if resolvedDef != nil {
				// If the reference was resolved, recursively unroll any
				// further references within the resolved definition itself.
				return unrollRecursive(resolvedDef, defs)
			} else {
				// If the reference could not be resolved (e.g., invalid format, not found),
				// return the original map with the unresolved $ref.
				// A more strict implementation might return an error here.
				return v
			}
		}

		// If no $ref is present, recursively process all fields in the map
		newMap := make(map[string]interface{})
		for key, val := range v {
			newMap[key] = unrollRecursive(val, defs)
		}
		return newMap

	case []interface{}:
		// If the current element is a slice (array), recursively process each element
		newSlice := make([]interface{}, len(v))
		for i, val := range v {
			newSlice[i] = unrollRecursive(val, defs)
		}
		return newSlice

	default:
		// If the element is a primitive type (string, number, boolean, null), return it as is
		return v
	}
}

// resolveRef is a helper function to find a definition by its $ref string.
// This implementation specifically handles internal references of the format
// "#/$defs/DefinitionName". It can be extended for other reference types.
func resolveRef(ref string, defs map[string]interface{}) interface{} {
	// Check if the reference is an internal pointer to the $defs section
	if strings.HasPrefix(ref, "#/$defs/") {
		// Extract the definition name from the reference string
		defName := strings.TrimPrefix(ref, "#/$defs/")
		// Look up the definition in the provided $defs map
		if def, ok := defs[defName]; ok {
			return def
		}
	}
	// Return nil if the reference format is not supported or the definition is not found
	return nil
}

func main() {
	// Example Usage with your provided schema:
	jsonSchema := `{
		"$defs": {
			"GoogleSearchRequest": {
				"properties": {
					"query": {
						"description": "The search query to send to Google",
						"examples": [
							"python programming",
							"machine learning basics"
						],
						"title": "Query",
						"type": "string"
					},
					"num_results": {
						"anyOf": [
							{
								"maximum": 10,
								"minimum": 1,
								"type": "integer"
							},
							{
								"type": "null"
							}
						],
						"default": 10,
						"description": "Number of results to return (max 10)",
						"examples": [
							5,
							10
						],
						"title": "Num Results"
					}
				},
				"required": [
					"query"
				],
				"title": "GoogleSearchRequest",
				"type": "object"
			}
		},
		"properties": {
			"request": {
				"$ref": "#/$defs/GoogleSearchRequest"
			}
		},
		"required": [
			"request"
		],
		"title": "searchArguments",
		"type": "object"
	}`

	var schema map[string]interface{}
	// Unmarshal the JSON string into a Go map
	err := json.Unmarshal([]byte(jsonSchema), &schema)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// Call the UnrollRefs function to process the schema
	unrolled, err := UnrollRefs(schema)
	if err != nil {
		fmt.Println("Error unrolling refs:", err)
		return
	}

	// Marshal the unrolled schema back into a JSON string for printing
	unrolledJSON, err := json.MarshalIndent(unrolled, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling unrolled JSON:", err)
		return
	}

	// Print the resulting unrolled JSON schema
	fmt.Println(string(unrolledJSON))
}
