package commands

import (
	"strings"
	"testing"
)

func TestProjectUsersListSendsPagination(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"project_users":[],"total_count":0,"page":2,"page_size":5,"total_pages":2}`,
		"project-users", "list", "--project", "p1", "--page", "2", "--page-size", "5")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if req.Query.Get("page") != "2" || req.Query.Get("page_size") != "5" {
		t.Errorf("query = %v, want page=2 page_size=5", req.Query)
	}
}

func TestProjectUsersCreateSendsUserEmail(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"project_user":{"id":"pu1"}}`,
		"project-users", "create", "--project", "p1", "--user-email", "oliver@example.com")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	body, _ := req.Body["project_user"].(map[string]any)
	if body["user_email"] != "oliver@example.com" {
		t.Errorf("project_user = %v, want user_email oliver@example.com", body)
	}
	if _, ok := body["user_id"]; ok {
		t.Errorf("project_user = %v, want no user_id", body)
	}
}

func TestProjectUsersCreateNeedsExactlyOneUser(t *testing.T) {
	cases := map[string][]string{
		"neither": {"project-users", "create", "--project", "p1"},
		"both":    {"project-users", "create", "--project", "p1", "--user-id", "u1", "--user-email", "oliver@example.com"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := runAgainstServer(t, `{}`, args...)
			if err == nil || !strings.Contains(err.Error(), "pass --user-id or --user-email") {
				t.Errorf("Execute() error = %v, want the exactly-one error", err)
			}
		})
	}
}

func TestProjectUsersDeleteEscapesPathSegments(t *testing.T) {
	req, _, err := runAgainstServer(t, `{}`, "project-users", "delete", "u/1", "--project", "p/1")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if req.Method != "DELETE" || !strings.HasSuffix(req.EscapedPath, "/projects/p%2F1/project-users/u%2F1") {
		t.Errorf("request = %s %s, want DELETE with escaped project and user segments", req.Method, req.EscapedPath)
	}
}
