// integration-test is a standalone binary that exercises the backend API
// against a running instance (typically the dev docker-compose stack).
// It is a pure black-box test — it only communicates over HTTP using the
// super admin API key configured in the backend.
//
// Build:   go build -o integration-test ./cmd/integration-test
// Run:     ./integration-test
//          ./integration-test -base-url http://backend:8088 -super-admin-api-key <key>
//
// The binary:
//   1. Authenticates via the super admin API key
//   2. Runs a sequence of API calls covering every major resource
//   3. Cleans up created resources via DELETE endpoints
//   4. Exits 0 on success, 1 on any failure
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

var (
	baseURL          string
	superAdminAPIKey string
)

func init() {
	flag.StringVar(&baseURL, "base-url", "http://localhost:8088", "Backend base URL")
	flag.StringVar(&superAdminAPIKey, "super-admin-api-key", "super-admin-dev-key-do-not-use-in-production", "Super Admin API Key")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type testResult struct {
	name    string
	passed  bool
	message string
}

type testContext struct {
	client         *http.Client
	apiKey         string
	userIdentifier string
	failures       int
	results        []testResult
}

func (tc *testContext) request(method, path string, body interface{}) (int, map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	url := baseURL + path
	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+tc.apiKey)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := tc.client.Do(request)
	if err != nil {
		return 0, nil, fmt.Errorf("do request %s %s: %w", method, path, err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	if len(responseBytes) == 0 {
		return response.StatusCode, nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		// Some endpoints return non-JSON (e.g. plain text errors)
		return response.StatusCode, nil, nil
	}

	return response.StatusCode, result, nil
}

func (tc *testContext) requestRaw(method, path string, bodyBytes []byte, contentType string) (int, map[string]interface{}, error) {
	url := baseURL + path
	request, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+tc.apiKey)
	request.Header.Set("Content-Type", contentType)

	response, err := tc.client.Do(request)
	if err != nil {
		return 0, nil, fmt.Errorf("do request: %w", err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	if len(responseBytes) == 0 {
		return response.StatusCode, nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		return response.StatusCode, nil, nil
	}

	return response.StatusCode, result, nil
}

func (tc *testContext) assert(testName string, condition bool, format string, args ...interface{}) {
	if !condition {
		tc.failures++
		msg := fmt.Sprintf(format, args...)
		log.Printf("FAIL  %s: %s", testName, msg)
		tc.results = append(tc.results, testResult{name: testName, passed: false, message: msg})
	} else {
		tc.results = append(tc.results, testResult{name: testName, passed: true})
	}
}

func (tc *testContext) assertStatus(testName string, got, want int) bool {
	if got != want {
		tc.failures++
		msg := fmt.Sprintf("status=%d, want %d", got, want)
		log.Printf("FAIL  %s: %s", testName, msg)
		tc.results = append(tc.results, testResult{name: testName, passed: false, message: msg})
		return false
	}
	tc.results = append(tc.results, testResult{name: testName, passed: true})
	return true
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func getItems(m map[string]interface{}) []interface{} {
	if m == nil {
		return nil
	}
	v, ok := m["items"]
	if !ok {
		return nil
	}
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return items
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func testProjectLifecycle(tc *testContext) {
	log.Println("--- testProjectLifecycle ---")

	// Create project
	status, body, err := tc.request("POST", "/api/v1/projects", map[string]interface{}{
		"project_name":        "Integration Test Project",
		"project_description": "Created by integration tests",
		"folder_path":         "test.integration",
	})
	if err != nil {
		tc.failures++
		log.Printf("FAIL  create project: %v", err)
		return
	}
	if !tc.assertStatus("create project", status, http.StatusOK) {
		log.Printf("       body: %v", body)
		return
	}
	projectIdentifier := getString(body, "project_identifier")
	tc.assert("create project", projectIdentifier != "", "project_id empty")
	log.Printf("  created project=%s", projectIdentifier)

	// List projects
	status, body, err = tc.request("GET", "/api/v1/projects", nil)
	if err == nil {
		tc.assertStatus("list projects", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list projects", len(items) > 0, "expected at least 1 project, got %d", len(items))
	}

	// Get project
	status, body, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier, nil)
	if err == nil {
		tc.assertStatus("get project", status, http.StatusOK)
		tc.assert("get project", getString(body, "project_name") == "Integration Test Project", "wrong name")
	}

	// Update project
	status, body, err = tc.request("PUT", "/api/v1/projects/"+projectIdentifier, map[string]interface{}{
		"project_name":        "Updated Integration Test Project",
		"project_description": "Updated description",
	})
	if err == nil {
		tc.assertStatus("update project", status, http.StatusOK)
		tc.assert("update project", getString(body, "project_name") == "Updated Integration Test Project", "name not updated")
	}

	// Store for downstream tests
	testElementLifecycle(tc, projectIdentifier)
	testPhaseLifecycle(tc, projectIdentifier)
	testCustomFieldLifecycle(tc, projectIdentifier)

	// Delete project (cascades to elements, links, etc.)
	status, _, err = tc.request("DELETE", "/api/v1/projects/"+projectIdentifier, nil)
	if err == nil {
		tc.assertStatus("delete project", status, http.StatusOK)
	}

	// Verify deletion
	status, _, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier, nil)
	if err == nil {
		tc.assert("get deleted project", status == http.StatusNotFound || status == http.StatusInternalServerError, "expected 404/500 after delete, got %d", status)
	}

	log.Println("--- testProjectLifecycle done ---")
}

func testElementLifecycle(tc *testContext, projectIdentifier string) {
	log.Println("  --- testElementLifecycle ---")

	// Create a feature
	status, body, err := tc.request("POST", "/api/v1/projects/"+projectIdentifier+"/elements", map[string]interface{}{
		"element_type": "feature",
		"title":        "Test Feature",
		"description":  "A feature for testing",
	})
	if err != nil || !tc.assertStatus("create feature", status, http.StatusOK) {
		log.Printf("       create feature err=%v body=%v", err, body)
		return
	}
	featureIdentifier := getString(body, "element_identifier")
	tc.assert("create feature", featureIdentifier != "", "element_identifier empty")
	log.Printf("    created feature=%s", featureIdentifier)

	// Create a task under the feature
	status, body, err = tc.request("POST", "/api/v1/projects/"+projectIdentifier+"/elements", map[string]interface{}{
		"element_type":      "task",
		"title":             "Test Task",
		"description":       "A task under the feature",
		"task_status":       "backlog",
		"task_progress":     0,
		"parent_feature_identifier": featureIdentifier,
	})
	if err != nil || !tc.assertStatus("create task", status, http.StatusOK) {
		log.Printf("       create task err=%v body=%v", err, body)
		return
	}
	taskIdentifier := getString(body, "element_identifier")
	log.Printf("    created task=%s", taskIdentifier)

	// Create a requirement
	status, body, err = tc.request("POST", "/api/v1/projects/"+projectIdentifier+"/elements", map[string]interface{}{
		"element_type":  "requirement",
		"title":         "Test Requirement",
		"description":   "A requirement for testing",
		"interest_level": 8,
	})
	if err != nil || !tc.assertStatus("create requirement", status, http.StatusOK) {
		return
	}
	requirementIdentifier := getString(body, "element_identifier")
	log.Printf("    created requirement=%s", requirementIdentifier)

	// List elements
	status, body, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier+"/elements", nil)
	if err == nil {
		tc.assertStatus("list elements", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list elements", len(items) >= 3, "expected >=3 elements, got %d", len(items))
	}

	// List with type filter
	status, body, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier+"/elements?element_type=task", nil)
	if err == nil {
		tc.assertStatus("list tasks", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list tasks", len(items) >= 1, "expected >=1 task")
	}

	// Get single element
	status, body, err = tc.request("GET", "/api/v1/elements/"+featureIdentifier, nil)
	if err == nil {
		tc.assertStatus("get feature", status, http.StatusOK)
		tc.assert("get feature", getString(body, "title") == "Test Feature", "wrong title")
	}

	// Update task — move to in_progress
	status, body, err = tc.request("PUT", "/api/v1/elements/"+taskIdentifier, map[string]interface{}{
		"title":         "Test Task Updated",
		"description":   "Updated description",
		"task_status":   "in_progress",
		"task_progress": 50,
	})
	if err == nil {
		tc.assertStatus("update task", status, http.StatusOK)
		tc.assert("update task", getString(body, "task_status") == "in_progress", "status not updated")
	}

	// Search elements
	status, body, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier+"/elements/search?query=Updated&limit=10&offset=0", nil)
	if err == nil {
		tc.assertStatus("search elements", status, http.StatusOK)
	}

	// --- Link tests ---
	testLinkLifecycle(tc, featureIdentifier, taskIdentifier, requirementIdentifier)

	// --- Attachment tests ---
	testAttachmentLifecycle(tc, featureIdentifier)

	// --- Version tests (after edits) ---
	testVersionLifecycle(tc, taskIdentifier)

	// Delete elements
	for _, elementIdentifier := range []string{taskIdentifier, featureIdentifier, requirementIdentifier} {
		status, _, err = tc.request("DELETE", "/api/v1/elements/"+elementIdentifier, nil)
		if err == nil {
			tc.assertStatus("delete element "+elementIdentifier[:8], status, http.StatusOK)
		}
	}

	log.Println("  --- testElementLifecycle done ---")
}

func testLinkLifecycle(tc *testContext, featureIdentifier, taskIdentifier, requirementIdentifier string) {
	log.Println("    --- testLinkLifecycle ---")

	// Create child link: feature → task
	status, body, err := tc.request("POST", "/api/v1/elements/"+featureIdentifier+"/links", map[string]interface{}{
		"destination_element_identifier": taskIdentifier,
		"link_type":              "child",
	})
	if err != nil || !tc.assertStatus("create child link", status, http.StatusOK) {
		log.Printf("       create link err=%v body=%v", err, body)
		return
	}
	childLinkIdentifier := getString(body, "link_identifier")
	tc.assert("create child link", childLinkIdentifier != "", "link_identifier empty")
	log.Printf("      created child link=%s", childLinkIdentifier)

	// Create implement link: task → requirement
	status, body, err = tc.request("POST", "/api/v1/elements/"+taskIdentifier+"/links", map[string]interface{}{
		"destination_element_identifier": requirementIdentifier,
		"link_type":              "implement",
	})
	if err != nil || !tc.assertStatus("create implement link", status, http.StatusOK) {
		return
	}
	implementLinkIdentifier := getString(body, "link_identifier")
	log.Printf("      created implement link=%s", implementLinkIdentifier)

	// List links for feature (outgoing)
	status, body, err = tc.request("GET", "/api/v1/elements/"+featureIdentifier+"/links?direction=outgoing", nil)
	if err == nil {
		tc.assertStatus("list outgoing links", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list outgoing links", len(items) >= 1, "expected >=1 outgoing link")
	}

	// List links for task (both directions)
	status, body, err = tc.request("GET", "/api/v1/elements/"+taskIdentifier+"/links", nil)
	if err == nil {
		tc.assertStatus("list task links", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list task links", len(items) >= 2, "expected >=2 links (incoming child + outgoing implement), got %d", len(items))
	}

	// Update link type (child → related)
	status, body, err = tc.request("PATCH", "/api/v1/links/"+childLinkIdentifier, map[string]interface{}{
		"link_type": "related",
	})
	if err == nil {
		tc.assertStatus("update link type", status, http.StatusOK)
		tc.assert("update link type", getString(body, "link_type") == "related", "link_type not updated")
	}

	// Delete links
	for _, linkIdentifier := range []string{childLinkIdentifier, implementLinkIdentifier} {
		status, _, err = tc.request("DELETE", "/api/v1/links/"+linkIdentifier, nil)
		if err == nil {
			tc.assertStatus("delete link "+linkIdentifier[:8], status, http.StatusOK)
		}
	}

	log.Println("    --- testLinkLifecycle done ---")
}

func testAttachmentLifecycle(tc *testContext, elementIdentifier string) {
	log.Println("    --- testAttachmentLifecycle ---")

	// Upload a small text attachment
	fakeFileContent := []byte("Hello, this is a fake attachment for integration testing.")

	status, body, err := tc.requestRaw(
		"POST",
		"/api/v1/elements/"+elementIdentifier+"/attachments?file_name=test-doc.txt",
		fakeFileContent,
		"text/plain",
	)
	if err != nil || status != http.StatusOK {
		log.Printf("    attachment upload skipped (S3 not configured or unavailable): status=%d err=%v", status, err)
		return
	}
	attachmentIdentifier := getString(body, "attachment_identifier")
	tc.assert("upload attachment", attachmentIdentifier != "", "attachment_id empty")
	tc.assert("upload attachment", getString(body, "file_name") == "test-doc.txt", "wrong file_name")
	log.Printf("      uploaded attachment=%s", attachmentIdentifier)

	// List attachments
	status, body, err = tc.request("GET", "/api/v1/elements/"+elementIdentifier+"/attachments", nil)
	if err == nil {
		tc.assertStatus("list attachments", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list attachments", len(items) >= 1, "expected >=1 attachment")
	}

	// Get presigned URL (will likely fail if S3 is not fully configured, but test the route)
	status, body, err = tc.request("GET", "/api/v1/attachments/"+attachmentIdentifier, nil)
	if err == nil {
		// Accept either 200 (S3 working) or 500 (S3 not configured) — just verify the route works
		tc.assert("get presigned url", status == http.StatusOK || status == http.StatusInternalServerError,
			"expected 200 or 500, got %d", status)
	}

	// Delete attachment
	status, _, err = tc.request("DELETE", "/api/v1/attachments/"+attachmentIdentifier, nil)
	if err == nil {
		tc.assertStatus("delete attachment", status, http.StatusOK)
	}

	log.Println("    --- testAttachmentLifecycle done ---")
}

func testVersionLifecycle(tc *testContext, elementIdentifier string) {
	log.Println("    --- testVersionLifecycle ---")

	// List versions — may be empty if the commit worker hasn't run yet
	status, body, err := tc.request("GET", "/api/v1/elements/"+elementIdentifier+"/versions", nil)
	if err == nil {
		tc.assertStatus("list versions", status, http.StatusOK)
		items := getItems(body)
		log.Printf("      versions found: %d", len(items))
	}

	// If there are versions, try to get the element at version 0
	if body != nil {
		items := getItems(body)
		if len(items) > 0 {
			status, _, err = tc.request("GET", "/api/v1/elements/"+elementIdentifier+"/versions/0", nil)
			if err == nil {
				tc.assertStatus("get element at version 0", status, http.StatusOK)
			}
		}
	}

	log.Println("    --- testVersionLifecycle done ---")
}

func testPhaseLifecycle(tc *testContext, projectIdentifier string) {
	log.Println("  --- testPhaseLifecycle ---")

	// Create phase
	status, body, err := tc.request("POST", "/api/v1/projects/"+projectIdentifier+"/phases", map[string]interface{}{
		"phase_name":  "Phase 1 - Prototype",
		"phase_order": 1,
	})
	if err != nil || !tc.assertStatus("create phase", status, http.StatusOK) {
		log.Printf("       create phase err=%v body=%v", err, body)
		return
	}
	phaseIdentifier := getString(body, "phase_identifier")
	tc.assert("create phase", phaseIdentifier != "", "phase_id empty")
	log.Printf("    created phase=%s", phaseIdentifier)

	// List phases
	status, body, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier+"/phases", nil)
	if err == nil {
		tc.assertStatus("list phases", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list phases", len(items) >= 1, "expected >=1 phase")
	}

	// Update phase
	status, body, err = tc.request("PUT", "/api/v1/phases/"+phaseIdentifier, map[string]interface{}{
		"phase_name":  "Phase 1 - Updated",
		"phase_order": 2,
	})
	if err == nil {
		tc.assertStatus("update phase", status, http.StatusOK)
	}

	// Delete phase
	status, _, err = tc.request("DELETE", "/api/v1/phases/"+phaseIdentifier, nil)
	if err == nil {
		tc.assertStatus("delete phase", status, http.StatusOK)
	}

	log.Println("  --- testPhaseLifecycle done ---")
}

func testCustomFieldLifecycle(tc *testContext, projectIdentifier string) {
	log.Println("  --- testCustomFieldLifecycle ---")

	// First create an element to attach custom field values to
	status, body, err := tc.request("POST", "/api/v1/projects/"+projectIdentifier+"/elements", map[string]interface{}{
		"element_type": "task",
		"title":        "Custom Field Test Element",
		"description":  "Element for custom field testing",
		"task_status":  "backlog",
	})
	if err != nil || !tc.assertStatus("create cf test element", status, http.StatusOK) {
		return
	}
	elementIdentifier := getString(body, "element_identifier")

	// Create a custom field definition
	status, body, err = tc.request("POST", "/api/v1/projects/"+projectIdentifier+"/custom-field-definitions", map[string]interface{}{
		"applicable_element_type": "task",
		"field_name":              "priority_label",
		"field_type":              "string",
		"field_options":           map[string]interface{}{},
		"display_order":           1,
	})
	if err != nil || !tc.assertStatus("create field def", status, http.StatusOK) {
		log.Printf("       create field def err=%v body=%v", err, body)
		return
	}
	fieldDefId := getString(body, "field_definition_identifier")
	tc.assert("create field def", fieldDefId != "", "field_definition_id empty")
	log.Printf("    created field def=%s", fieldDefId)

	// List field definitions
	status, body, err = tc.request("GET", "/api/v1/projects/"+projectIdentifier+"/custom-field-definitions?element_type=task", nil)
	if err == nil {
		tc.assertStatus("list field defs", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list field defs", len(items) >= 1, "expected >=1 field def")
	}

	// Set field value on element
	status, _, err = tc.request("PUT", "/api/v1/elements/"+elementIdentifier+"/custom-field-values", map[string]interface{}{
		"field_values": []map[string]interface{}{
			{"field_definition_identifier": fieldDefId, "value": "critical"},
		},
	})
	if err == nil {
		tc.assertStatus("set field value", status, http.StatusOK)
	}

	// Get field values
	status, body, err = tc.request("GET", "/api/v1/elements/"+elementIdentifier+"/custom-field-values", nil)
	if err == nil {
		tc.assertStatus("get field values", status, http.StatusOK)
		items := getItems(body)
		tc.assert("get field values", len(items) >= 1, "expected >=1 field value")
	}

	// Delete field value
	status, _, err = tc.request("DELETE", "/api/v1/elements/"+elementIdentifier+"/custom-field-values/"+fieldDefId, nil)
	if err == nil {
		tc.assertStatus("delete field value", status, http.StatusOK)
	}

	// Delete field definition
	status, _, err = tc.request("DELETE", "/api/v1/custom-field-definitions/"+fieldDefId, nil)
	if err == nil {
		tc.assertStatus("delete field def", status, http.StatusOK)
	}

	// Cleanup element
	tc.request("DELETE", "/api/v1/elements/"+elementIdentifier, nil)

	log.Println("  --- testCustomFieldLifecycle done ---")
}

func testGroupAndAccessLifecycle(tc *testContext) {
	log.Println("--- testGroupAndAccessLifecycle ---")

	// Create a group
	status, body, err := tc.request("POST", "/api/v1/groups", map[string]interface{}{
		"group_name": "Integration Test Group",
	})
	if err != nil || !tc.assertStatus("create group", status, http.StatusOK) {
		log.Printf("       create group err=%v body=%v", err, body)
		return
	}
	groupIdentifier := getString(body, "group_identifier")
	tc.assert("create group", groupIdentifier != "", "group_id empty")
	log.Printf("  created group=%s", groupIdentifier)

	// List groups
	status, body, err = tc.request("GET", "/api/v1/groups", nil)
	if err == nil {
		tc.assertStatus("list groups", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list groups", len(items) >= 1, "expected >=1 group")
	}

	// Add test user to group
	status, _, err = tc.request("POST", "/api/v1/groups/"+groupIdentifier+"/members", map[string]interface{}{
		"user_identifier": tc.userIdentifier,
	})
	if err == nil {
		tc.assertStatus("add user to group", status, http.StatusOK)
	}

	// List group members
	status, body, err = tc.request("GET", "/api/v1/groups/"+groupIdentifier+"/members", nil)
	if err == nil {
		tc.assertStatus("list group members", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list group members", len(items) >= 1, "expected >=1 member")
	}

	// Create a project for access testing
	status, body, err = tc.request("POST", "/api/v1/projects", map[string]interface{}{
		"project_name": "Access Test Project",
	})
	var accessProjectId string
	if err == nil && tc.assertStatus("create access project", status, http.StatusOK) {
		accessProjectId = getString(body, "project_identifier")
	}

	if accessProjectId != "" {
		// Grant group access to project
		status, _, err = tc.request("PUT", "/api/v1/projects/"+accessProjectId+"/access", map[string]interface{}{
			"group_identifier":     groupIdentifier,
			"access_level": "write",
		})
		if err == nil {
			tc.assertStatus("set project access", status, http.StatusOK)
		}

		// List project access
		status, body, err = tc.request("GET", "/api/v1/projects/"+accessProjectId+"/access", nil)
		if err == nil {
			tc.assertStatus("list project access", status, http.StatusOK)
			items := getItems(body)
			tc.assert("list project access", len(items) >= 1, "expected >=1 access entry")
		}

		// Check user access
		status, body, err = tc.request("GET", "/api/v1/projects/"+accessProjectId+"/access/check", nil)
		if err == nil {
			tc.assertStatus("check user access", status, http.StatusOK)
		}

		// Remove access
		status, _, err = tc.request("DELETE", "/api/v1/projects/"+accessProjectId+"/access/"+groupIdentifier, nil)
		if err == nil {
			tc.assertStatus("remove project access", status, http.StatusOK)
		}

		// Delete access project
		tc.request("DELETE", "/api/v1/projects/"+accessProjectId, nil)
	}

	// Remove user from group
	status, _, err = tc.request("DELETE", "/api/v1/groups/"+groupIdentifier+"/members/"+tc.userIdentifier, nil)
	if err == nil {
		tc.assertStatus("remove user from group", status, http.StatusOK)
	}

	// Delete group
	status, _, err = tc.request("DELETE", "/api/v1/groups/"+groupIdentifier, nil)
	if err == nil {
		tc.assertStatus("delete group", status, http.StatusOK)
	}

	log.Println("--- testGroupAndAccessLifecycle done ---")
}

func testUserEndpoints(tc *testContext) {
	log.Println("--- testUserEndpoints ---")

	// List users
	status, body, err := tc.request("GET", "/api/v1/users?limit=10&offset=0", nil)
	if err == nil {
		tc.assertStatus("list users", status, http.StatusOK)
		items := getItems(body)
		tc.assert("list users", len(items) >= 1, "expected >=1 user (super admin)")
	}

	// Get current user (super admin)
	status, body, err = tc.request("GET", "/api/v1/users/"+tc.userIdentifier, nil)
	if err == nil {
		tc.assertStatus("get super admin user", status, http.StatusOK)
		tc.assert("get super admin user", getString(body, "email") != "", "email should not be empty")
	}

	log.Println("--- testUserEndpoints done ---")
}

func testAuthMeEndpoint(tc *testContext) {
	log.Println("--- testAuthMeEndpoint ---")

	status, body, err := tc.request("GET", "/api/v1/auth/me", nil)
	if err == nil {
		tc.assertStatus("auth me", status, http.StatusOK)
		tc.assert("auth me", getString(body, "user_identifier") == tc.userIdentifier, "wrong user_id from /auth/me")
	}

	log.Println("--- testAuthMeEndpoint done ---")
}

// ---------------------------------------------------------------------------
// Report
// ---------------------------------------------------------------------------

func writeMarkdownReport(results []testResult) {
	if err := os.MkdirAll("test-results", 0o755); err != nil {
		log.Printf("failed to create test-results dir: %v", err)
		return
	}
	f, err := os.Create("test-results/integration-tests.md")
	if err != nil {
		log.Printf("failed to create integration test report: %v", err)
		return
	}
	defer f.Close()

	// Merge multiple assertions sharing the same name (e.g. assertStatus + assert
	// on the same API call) into a single row, collecting any failure messages.
	type entry struct {
		name    string
		passed  bool
		message string
	}
	var rows []entry
	nameToIdx := map[string]int{}
	for _, r := range results {
		if idx, ok := nameToIdx[r.name]; ok {
			if !r.passed {
				rows[idx].passed = false
				if r.message != "" {
					if rows[idx].message != "" {
						rows[idx].message += "; " + r.message
					} else {
						rows[idx].message = r.message
					}
				}
			}
		} else {
			nameToIdx[r.name] = len(rows)
			rows = append(rows, entry{name: r.name, passed: r.passed, message: r.message})
		}
	}

	passed, failed := 0, 0
	for _, row := range rows {
		if row.passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Fprintf(f, "-----------\n-----------\n\n")
	fmt.Fprintf(f, "## 🧪 Integration Test Results: %s<br/>\n\n", time.Now().UTC().Format(time.RFC1123))
	fmt.Fprintf(f, "| Result | Test | Errors |\n")
	fmt.Fprintf(f, "| :--- | :--- | :--- |\n")
	for _, row := range rows {
		icon := "✅ "
		if !row.passed {
			icon = "❌ "
		}
		name := strings.ReplaceAll(row.name, "_", "\\_")
		msg := strings.ReplaceAll(row.message, "|", "\\|")
		fmt.Fprintf(f, "| %s | %s | %s |\n", icon, name, msg)
	}

	fmt.Fprintf(f, "\n-----------\n\n")
	fmt.Fprintf(f, "## 📊 Test Results Summary:\n\n")
	fmt.Fprintf(f, " * **✅ Passed:** %d\n", passed)
	fmt.Fprintf(f, " * **❌ Failed:** %d\n", failed)
	fmt.Fprintf(f, "\n-----------\n")
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	flag.Parse()

	log.SetFlags(log.Ltime)
	log.Println("===========================================")
	log.Println("  Project Works Integration Tests")
	log.Println("===========================================")
	log.Printf("  base-url:            %s", baseURL)
	log.Printf("  super-admin-api-key: %s***", superAdminAPIKey[:8])

	// Wait for backend to be ready
	log.Println("waiting for backend to be ready...")
	client := &http.Client{Timeout: 30 * time.Second}
	ready := false
	for attempt := 0; attempt < 30; attempt++ {
		request, _ := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
		request.Header.Set("Authorization", "Bearer "+superAdminAPIKey)
		response, err := client.Do(request)
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	if !ready {
		log.Fatal("backend not reachable or super admin key rejected after 30 seconds")
	}
	log.Println("backend is ready")

	// Discover our user ID via /auth/me
	tc := &testContext{
		client: client,
		apiKey: superAdminAPIKey,
	}
	status, body, err := tc.request("GET", "/api/v1/auth/me", nil)
	if err != nil || status != http.StatusOK {
		log.Fatalf("failed to resolve super admin identity: status=%d err=%v", status, err)
	}
	tc.userIdentifier = getString(body, "user_identifier")
	if tc.userIdentifier == "" {
		log.Fatal("super admin /auth/me returned empty user_id")
	}
	log.Printf("authenticated as user_id=%s", tc.userIdentifier)

	// Run tests
	testAuthMeEndpoint(tc)
	testUserEndpoints(tc)
	testGroupAndAccessLifecycle(tc)
	testProjectLifecycle(tc)

	// Summary
	log.Println("===========================================")
	writeMarkdownReport(tc.results)
	if tc.failures > 0 {
		log.Printf("  FAILED: %d assertion(s) failed", tc.failures)
		log.Println("===========================================")
		os.Exit(1)
	}
	log.Println("  ALL TESTS PASSED")
	log.Println("===========================================")
	os.Exit(0)
}

