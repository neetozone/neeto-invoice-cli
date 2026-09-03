package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neetozone/neeto-invoice-cli/internal/output"
)

func TestReadJSONFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	if err := os.WriteFile(path, []byte(`{"name":"test","count":42}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := readJSONFile(path)
	if err != nil {
		t.Fatalf("readJSONFile() error = %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["count"] != float64(42) {
		t.Errorf("count = %v, want 42", result["count"])
	}
}

func TestReadJSONFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{not valid`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := readJSONFile(path)
	if err == nil {
		t.Error("readJSONFile() expected error for invalid JSON")
	}
}

func TestReadJSONFile_NotFound(t *testing.T) {
	_, err := readJSONFile("/nonexistent/file.json")
	if err == nil {
		t.Error("readJSONFile() expected error for missing file")
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	return buf.String()
}

var timeEntriesPaginationKeys = inlinePaginationKeys{
	TotalRecords: "total_count",
	TotalPages:   "total_pages",
	CurrentPage:  "page",
	PageSize:     "page_size",
}

func jsonEqual(t *testing.T, got, want string) bool {
	t.Helper()

	var g, w interface{}
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		return false
	}
	return reflect.DeepEqual(g, w)
}

func TestPrintListWithInlinePagination(t *testing.T) {
	tests := []struct {
		name             string
		data             string
		resourceKey      string
		keys             inlinePaginationKeys
		wantData         string
		wantEnvelope     bool
		wantNoPagination bool
	}{
		{
			name:         "lifts inline paging keys into the envelope",
			data:         `{"time_entries":[{"id":"t1"}],"total_count":42,"page":2,"page_size":10,"total_pages":5,"next_page":3,"prev_page":1}`,
			resourceKey:  "time_entries",
			keys:         timeEntriesPaginationKeys,
			wantData:     `[{"id":"t1"}]`,
			wantEnvelope: true,
		},
		{
			name:             "falls back to raw data when resource key is missing",
			data:             `{"unrelated":"value"}`,
			resourceKey:      "time_entries",
			keys:             timeEntriesPaginationKeys,
			wantData:         `{"unrelated":"value"}`,
			wantEnvelope:     true,
			wantNoPagination: true,
		},
		{
			name:        "falls back to raw data on malformed JSON",
			data:        `not json`,
			resourceKey: "time_entries",
			keys:        timeEntriesPaginationKeys,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.ForceJSON = true
			defer func() { output.ForceJSON = false }()

			out := captureStdout(t, func() {
				printListWithInlinePagination(json.RawMessage(tt.data), tt.resourceKey, tt.keys, nil)
			})

			if !tt.wantEnvelope {
				if bytes.Contains([]byte(out), []byte("current_page_number")) {
					t.Errorf("output = %q, want no pagination block", out)
				}
				if len(bytes.TrimSpace([]byte(out))) != 0 {
					t.Errorf("output = %q, want empty output for malformed JSON", out)
				}
				return
			}

			var envelope output.Envelope
			if err := json.Unmarshal([]byte(out), &envelope); err != nil {
				t.Fatalf("output is not a valid envelope: %v\noutput: %s", err, out)
			}

			if !jsonEqual(t, string(envelope.Data), tt.wantData) {
				t.Errorf("data = %s, want %s", envelope.Data, tt.wantData)
			}

			if tt.wantNoPagination {
				if len(envelope.Pagination) != 0 {
					t.Errorf("pagination = %s, want none", envelope.Pagination)
				}
				return
			}

			var pagination struct {
				TotalRecords      int `json:"total_records"`
				TotalPages        int `json:"total_pages"`
				CurrentPageNumber int `json:"current_page_number"`
				PageSize          int `json:"page_size"`
			}
			if err := json.Unmarshal(envelope.Pagination, &pagination); err != nil {
				t.Fatalf("pagination is not valid JSON: %v", err)
			}
			if pagination.TotalRecords != 42 {
				t.Errorf("total_records = %d, want 42", pagination.TotalRecords)
			}
			if pagination.TotalPages != 5 {
				t.Errorf("total_pages = %d, want 5", pagination.TotalPages)
			}
			if pagination.CurrentPageNumber != 2 {
				t.Errorf("current_page_number = %d, want 2", pagination.CurrentPageNumber)
			}
			if pagination.PageSize != 10 {
				t.Errorf("page_size = %d, want 10", pagination.PageSize)
			}

			if bytes.Contains([]byte(out), []byte("next_page")) || bytes.Contains([]byte(out), []byte("prev_page")) {
				t.Errorf("rendered output still contains inline paging keys: %s", out)
			}
		})
	}
}

func TestIntFromRaw(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
		want int
	}{
		{name: "integer", raw: json.RawMessage(`42`), want: 42},
		{name: "null", raw: json.RawMessage(`null`), want: 0},
		{name: "missing", raw: nil, want: 0},
		{name: "not a number", raw: json.RawMessage(`"oops"`), want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intFromRaw(tt.raw); got != tt.want {
				t.Errorf("intFromRaw(%s) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}
