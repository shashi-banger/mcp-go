package client

import (
	"encoding/json"
	"testing"
)

var testResponseMsg = `
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "search",
        "description": "\n    Perform a Google search and return the results.\n\n    Args:\n        request: A search request containing:\n            - query (str): The search query to send to Google\n            - num_results (int, optional): Number of results to return (default: 10, max: 10)\n\n    Returns:\n        GoogleSearchResponse: A list of search results containing titles, links, and snippets\n    ",
        "inputSchema": {
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
        }
      }
    ]
  }
}
`

func TestListToolsPP(t *testing.T) {
	response := json.RawMessage(testResponseMsg)
	result, err := ListToolsPP(&response)
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}

	// Parse the result into a map[string]interface{}
	var resultMap map[string]interface{}
	err = json.Unmarshal(*result, &resultMap)
	if err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	for _, tool := range resultMap["result"].(map[string]interface{})["tools"].([]interface{}) {
		toolMap := tool.(map[string]interface{})
		inputSchema, ok := toolMap["inputSchema"].(map[string]interface{})

		if !ok {
			t.Fatalf("inputSchema is not a map[string]interface{}")
		}

		_, ok = inputSchema["$defs"].(map[string]interface{})
		if ok {
			t.Fatalf("$defs is not a map[string]interface{}")
		}
	}

}
