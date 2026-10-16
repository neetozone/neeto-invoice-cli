package commands

import "testing"

func TestClientsListSendsPaginationAndFilters(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"clients":[],"total_count":0,"page":3,"page_size":10,"total_pages":3}`,
		"clients", "list", "--name", "Acme", "--status", "active", "--page", "3", "--page-size", "10")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for key, value := range map[string]string{"name": "Acme", "status": "active", "page": "3", "page_size": "10"} {
		if got := req.Query.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
}
