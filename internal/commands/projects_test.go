package commands

import "testing"

func TestProjectsListSendsNameAndPagination(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"projects":[],"total_count":0,"page":2,"page_size":20,"total_pages":2}`,
		"projects", "list", "--name", "Website", "--client-id", "abc123", "--page", "2", "--page-size", "20")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for key, value := range map[string]string{"name": "Website", "client_id": "abc123", "page": "2", "page_size": "20"} {
		if got := req.Query.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
}
