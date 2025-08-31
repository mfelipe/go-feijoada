package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	hcconfig "github.com/tavsec/gin-healthcheck/config"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/modules/valkey"

	"github.com/mfelipe/go-feijoada/utils/testcontainers"
)

var (
	baseUrl = "http://127.0.0.1:8080"
)

var tc *testContext

type testContext struct {
	redisContainer  *redis.RedisContainer
	valkeyContainer *valkey.ValkeyContainer
}

func (tc *testContext) shutdown() {
	var err error
	if tc.redisContainer != nil {
		if terr := tc.redisContainer.Terminate(context.Background()); terr != nil {
			err = errors.Join(err, terr)
		}
	}
	if tc.valkeyContainer != nil {
		if terr := tc.valkeyContainer.Terminate(context.Background()); terr != nil {
			err = errors.Join(err, terr)
		}
	}

	if err != nil {
		panic(err)
	}
}

func setupTestContext() (*testContext, error) {
	ctx := context.Background()

	redisContainer := testcontainers.StartRedis(ctx)
	//valkeyContainer := testcontainers.StartValkey(ctx)

	redisAddress, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("setting SR_REPOSITORY_REDIS_ADDRESS as %s\n", redisAddress)
	_ = os.Setenv("SR_REPOSITORY_REDIS_ADDRESS", redisAddress)
	_ = os.Setenv("SR_REPOSITORY_REDIS_CLIENTNAME", "schema-repository-test")

	tc = &testContext{
		redisContainer: redisContainer,
		//valkeyContainer: valkeyContainer,
	}

	// Overkill for experimentation
	var run chan error
	var startTimeout <-chan time.Time

	go func() {
		run = startServer()
	}()
	go func() {
		startTimeout = time.After(20 * time.Second)
	}()

	// TODO: Get from config
	healthCheckURL := baseUrl + hcconfig.DefaultConfig().HealthPath

	for {
		select {
		case <-run:
			return tc, errors.New("server exited too soon")
		case <-startTimeout:
			return tc, errors.New("server start timed out")
		case <-time.After(time.Second):
			resp, err := http.Get(healthCheckURL)
			if err != nil {
				log.Err(err).Msg("failed to check server health status")
				continue
			}
			if resp.StatusCode == http.StatusOK {
				log.Info().Msg("server is ready")
				return tc, nil
			}
		}
	}
}

// Test schemas
var (
	validSchemaV1            = json.RawMessage(`{"type": "object", "properties": {"name": {"type": "string"}}}`)
	validCompatibleSchemaV12 = json.RawMessage(`{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}}`)
	invalidJSONSchema        = json.RawMessage(`{"type": "object", "properties": {"name": {"type": "string"`) // Missing closing brace
)

func TestMain(m *testing.M) {
	var err error
	defer func() {
		if tc != nil {
			tc.shutdown()
		}
	}()
	if tc, err = setupTestContext(); err != nil {
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func Test_AddSchema(t *testing.T) {
	tests := []struct {
		name           string
		schemaName     string
		schemaVersion  string
		schemaJSON     json.RawMessage
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid schema v1",
			schemaName:     "user",
			schemaVersion:  "1.0.0",
			schemaJSON:     validSchemaV1,
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "Valid schema v1.2 compatible",
			schemaName:     "user",
			schemaVersion:  "1.2.0",
			schemaJSON:     validCompatibleSchemaV12,
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "Invalid JSON schema",
			schemaName:     "user",
			schemaVersion:  "1.0.0",
			schemaJSON:     invalidJSONSchema,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Invalid schemaVersion format",
			schemaName:     "user",
			schemaVersion:  "invalid",
			schemaJSON:     validSchemaV1,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/schemas/%s/%s", baseUrl, tt.schemaName, tt.schemaVersion)
			body := map[string]json.RawMessage{"schema": tt.schemaJSON}
			jsonBody, _ := json.Marshal(body)

			resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonBody)))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer closeBody(resp)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectError {
				var response map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil && resp.StatusCode != http.StatusOK {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if _, hasError := response["error"]; !hasError {
					t.Error("Expected error in response, but got none")
				}
			} else {
				if resp.ContentLength != 0 {
					t.Error("Expected no schema in response, but it wasn't empty")
				}
			}
		})
	}
}

func Test_GetSchema(t *testing.T) {
	// First create a schema to test retrieval
	func() {
		schemaName := "product"
		version := "1.0.0"
		createURL := fmt.Sprintf("%s/schemas/%s/%s", baseUrl, schemaName, version)
		createBody := map[string]json.RawMessage{"schema": validSchemaV1}
		jsonBody, _ := json.Marshal(createBody)

		resp, err := http.Post(createURL, "application/json", strings.NewReader(string(jsonBody)))
		if err != nil || resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to create test schema: %v", err)
		}
	}()

	tests := []struct {
		name           string
		schemaName     string
		version        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Get existing schema",
			schemaName:     "product",
			version:        "1.0.0",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Get non-existing schema",
			schemaName:     "nonexistent",
			version:        "1.0.0",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Get with invalid schemaVersion",
			schemaName:     "product",
			version:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/schemas/%s/%s", baseUrl, tt.schemaName, tt.version)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer closeBody(resp)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var response map[string]interface{}
			if err = json.NewDecoder(resp.Body).Decode(&response); err != nil && resp.StatusCode != http.StatusOK {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.expectError {
				if _, hasError := response["error"]; !hasError {
					t.Error("Expected error in response, but got none")
				}
			} else {
				if _, hasContent := response["schema"]; !hasContent {
					t.Error("Expected schema in response, but got none")
				}
			}
		})
	}
}

func Test_DeleteSchema(t *testing.T) {
	// First create a schema to test deletion
	func() {
		schemaName := "toDelete"
		version := "1.0.0"
		createURL := fmt.Sprintf("%s/schemas/%s/%s", baseUrl, schemaName, version)
		createBody := map[string]json.RawMessage{"schema": validSchemaV1}
		jsonBody, _ := json.Marshal(createBody)

		resp, err := http.Post(createURL, "application/json", strings.NewReader(string(jsonBody)))
		if err != nil || resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to create test schema: %v", err)
		}
	}()

	tests := []struct {
		name           string
		schemaName     string
		version        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Delete existing schema",
			schemaName:     "toDelete",
			version:        "1.0.0",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Delete non-existing schema",
			schemaName:     "nonexistent",
			version:        "1.0.0",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Delete with invalid schemaVersion",
			schemaName:     "toDelete",
			version:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/schemas/%s/%s", baseUrl, tt.schemaName, tt.version)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer closeBody(resp)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectError {
				var response map[string]interface{}
				if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if _, hasError := response["error"]; !hasError {
					t.Error("Expected error in response, but got none")
				}
			}
		})
	}
}

func closeBody(body *http.Response) {
	if body != nil && body.Body != nil {
		_ = body.Body.Close()
	}
}
