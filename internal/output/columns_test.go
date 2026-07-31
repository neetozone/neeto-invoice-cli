package output

import (
	"encoding/json"
	"reflect"
	"testing"
)

func sample(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("bad sample json: %v", err)
	}
	return m
}

func TestPickColumnsPutsUsefulFieldsFirst(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    []string
	}{
		{
			name: "time entry",
			payload: `{"id":"1","user_id":"2","task_id":"3","hours":2.5,"recorded_on":"2026-07-14",
				"notes":"n","status":"submitted","created_at":"t","updated_at":"t","user_name":"Oliver Smith",
				"user_email":"oliver@example.com","task_name":"Design","project_id":"4",
				"project_name":"Website Redesign","client_id":"5","client_name":"Acme Corp"}`,
			want: []string{"id", "user_name", "user_email", "recorded_on", "hours", "task_name", "project_name"},
		},
		{
			name: "client",
			payload: `{"id":"1","identifier":"7c1f5e2a9b","name":"Acme Corp","secondary_name":"Acme Holdings",
				"status":"active","currency":"USD","default_invoice_due_in_days":30}`,
			want: []string{"id", "identifier", "name", "status", "currency", "default_invoice_due_in_days", "secondary_name"},
		},
		{
			name: "project",
			payload: `{"id":"1","identifier":"89d1b47d","name":"Website Redesign","client_id":"2",
				"status":"active","requires_daily_entry":false}`,
			want: []string{"id", "identifier", "name", "status", "client_id", "requires_daily_entry"},
		},
		{
			name: "project user",
			payload: `{"id":"1","user_id":"2","role":"project_manager","hourly_rate":120.0,
				"name":"Oliver Smith","email":"oliver@example.com","time_zone":"Asia/Kolkata"}`,
			want: []string{"id", "name", "email", "role", "hourly_rate", "time_zone", "user_id"},
		},
		{
			name: "forced pto entry",
			payload: `{"id":"1","user_email":"oliver@example.com","user_name":"Oliver Smith",
				"recorded_on":"2026-07-04","hours":8.0}`,
			want: []string{"id", "user_name", "user_email", "recorded_on", "hours"},
		},
		{
			name: "monthly pto row",
			payload: `{"user_email":"oliver@example.com","user_name":"Oliver Smith","expected_hours":176.0,
				"hours_worked":168.0,"pto_earned":8.0,"pto_used":8.0,"previous_month_pto":4.0,
				"net_pto":4.0,"salary_deduction":0.0,"locked":false}`,
			want: []string{"user_name", "user_email", "pto_earned", "pto_used", "net_pto", "expected_hours", "hours_worked"},
		},
		{
			name: "payroll summary row",
			payload: `{"user_name":"Oliver Smith","user_email":"oliver@example.com","working_days":22,
				"present_days":21,"lop_days":1}`,
			want: []string{"user_name", "user_email", "working_days", "lop_days", "present_days"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := pickColumns(sample(t, tc.payload))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("pickColumns()\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestPickColumnsNeverExceedsMaxTableColumns(t *testing.T) {
	payload := sample(t, `{"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8,"i":9,"j":10}`)
	if got := len(pickColumns(payload)); got > maxTableColumns {
		t.Errorf("pickColumns returned %d columns, want at most %d", got, maxTableColumns)
	}
}

func TestPickColumnsHandlesEmptyPayload(t *testing.T) {
	if got := pickColumns(map[string]interface{}{}); len(got) != 0 {
		t.Errorf("pickColumns(empty) = %v, want no columns", got)
	}
}

func TestPickColumnsWithNoPriorityMatchesFallsBackToSortedScalars(t *testing.T) {
	payload := sample(t, `{"zebra":1,"apple":2,"mango":3}`)
	want := []string{"apple", "mango", "zebra"}
	if got := pickColumns(payload); !reflect.DeepEqual(got, want) {
		t.Errorf("pickColumns()\n got: %v\nwant: %v", got, want)
	}
}

func TestPickColumnsSkipsNonScalars(t *testing.T) {
	payload := sample(t, `{"name":"Acme","tags":["a","b"],"client":{"id":"1"}}`)
	got := pickColumns(payload)
	for _, col := range got {
		if col == "tags" || col == "client" {
			t.Errorf("pickColumns included non-scalar %q in %v", col, got)
		}
	}
}
