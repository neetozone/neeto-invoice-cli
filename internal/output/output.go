package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alpkeskin/gotoon"
	"golang.org/x/term"
)

type Breadcrumb struct {
	Label   string `json:"label"`
	Command string `json:"command"`
}

type Envelope struct {
	Data        json.RawMessage `json:"data"`
	Breadcrumbs []Breadcrumb    `json:"breadcrumbs,omitempty"`
	Pagination  json.RawMessage `json:"pagination,omitempty"`
}

var (
	ForceJSON bool
	QuietMode bool
	ToonMode  bool
)

// priorityFields controls which columns appear in tables and their order.
var priorityFields = []string{
	"sid", "name", "title", "email", "status", "kind", "type",
	"slug", "duration", "spot", "disabled", "default",
	"host_name", "host_email", "starts_at", "ends_at", "time_zone",
	"event", "amount", "currency", "wday", "start_time", "end_time",
	"date", "day",
}

const (
	maxTableColumns = 7
	minColWidth     = 6
	colPadding      = 3
)

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func UseJSON() bool {
	return ForceJSON || QuietMode || !IsTTY()
}

func Print(data json.RawMessage, breadcrumbs []Breadcrumb) {
	if ToonMode {
		printToon(data)
		return
	}

	if QuietMode {
		fmt.Println(string(data))
		return
	}

	if UseJSON() {
		printEnvelope(data, breadcrumbs, nil)
		return
	}

	printPretty(data)
	printBreadcrumbs(breadcrumbs)
}

func PrintWithPagination(data json.RawMessage, pagination json.RawMessage, breadcrumbs []Breadcrumb) {
	if ToonMode {
		printToon(data)
		printToonPagination(pagination)
		return
	}

	if QuietMode {
		fmt.Println(string(data))
		return
	}

	if UseJSON() {
		printEnvelope(data, breadcrumbs, pagination)
		return
	}

	printPretty(data)
	printPaginationSummary(pagination)
	printBreadcrumbs(breadcrumbs)
}

func PrintMessage(msg string) {
	if QuietMode {
		fmt.Println("success")
		return
	}

	if ToonMode {
		fmt.Println(msg)
		return
	}

	if UseJSON() {
		envelope := map[string]string{"message": msg}
		data, _ := json.Marshal(envelope)
		fmt.Println(string(data))
	} else {
		fmt.Println(msg)
	}
}

// PrintQuiet prints a minimal one-line summary of a resource suitable for
// action commands (create/update) where the full response body is not needed.
// In quiet mode it prints only the sid (or id, or name) so the caller gets
// just the identifier. In other modes it falls through to Print.
func PrintQuiet(data json.RawMessage, breadcrumbs []Breadcrumb) {
	if QuietMode {
		if id := extractIdentifier(data); id != "" {
			fmt.Println(id)
			return
		}
	}

	Print(data, breadcrumbs)
}

func extractIdentifier(data json.RawMessage) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	if len(raw) == 1 {
		for _, v := range raw {
			var inner map[string]json.RawMessage
			if json.Unmarshal(v, &inner) == nil {
				raw = inner
			}
		}
	}

	for _, key := range []string{"sid", "id", "name"} {
		if v, ok := raw[key]; ok {
			var s string
			if json.Unmarshal(v, &s) == nil {
				return s
			}
		}
	}

	return ""
}

func printEnvelope(data json.RawMessage, breadcrumbs []Breadcrumb, pagination json.RawMessage) {
	envelope := Envelope{
		Data:        data,
		Breadcrumbs: breadcrumbs,
		Pagination:  pagination,
	}
	out, _ := json.MarshalIndent(envelope, "", "  ")
	fmt.Println(string(out))
}

func printPretty(data json.RawMessage) {
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) == 0 {
			fmt.Println("No records found.")
			return
		}
		printTable(arr, 0)
		return
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		if inner, ok := singleNestedObject(data); ok {
			printObject(inner, 1)
			return
		}
		printObject(data, 1)
		return
	}

	out, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(out))
}

func printTable(rows []map[string]interface{}, indent int) {
	prefix := indentPrefix(indent)

	cols := pickColumns(rows[0])
	if len(cols) == 0 {
		printIndentedJSON(rows, indent)
		return
	}

	headers := make([]string, len(cols))
	for i, col := range cols {
		headers[i] = formatHeader(col)
	}

	grid := make([][]string, len(rows))
	for i, row := range rows {
		grid[i] = make([]string, len(cols))
		for j, col := range cols {
			grid[i][j] = formatValue(row[col])
		}
	}

	widths := calculateWidths(headers, grid, indent)
	pad := strings.Repeat(" ", colPadding)

	fmt.Print(prefix)
	for i, h := range headers {
		if i > 0 {
			fmt.Print(pad)
		}
		fmt.Printf("%-*s", widths[i], truncate(h, widths[i]))
	}
	fmt.Println()

	fmt.Print(prefix)
	for i, w := range widths {
		if i > 0 {
			fmt.Print(pad)
		}
		fmt.Print(strings.Repeat("─", w))
	}
	fmt.Println()

	for _, row := range grid {
		fmt.Print(prefix)
		for i, val := range row {
			if i > 0 {
				fmt.Print(pad)
			}
			fmt.Printf("%-*s", widths[i], truncate(val, widths[i]))
		}
		fmt.Println()
	}
}

func pickColumns(sample map[string]interface{}) []string {
	scalars := map[string]bool{}
	for k, v := range sample {
		if isScalar(v) {
			scalars[k] = true
		}
	}

	var cols []string
	used := map[string]bool{}

	for _, f := range priorityFields {
		if scalars[f] && !used[f] && len(cols) < maxTableColumns {
			cols = append(cols, f)
			used[f] = true
		}
	}

	var remaining []string
	for k := range scalars {
		if !used[k] {
			remaining = append(remaining, k)
		}
	}
	sort.Strings(remaining)
	for _, k := range remaining {
		if len(cols) >= maxTableColumns {
			break
		}
		cols = append(cols, k)
	}

	return cols
}

func calculateWidths(headers []string, grid [][]string, indent int) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range grid {
		for i, val := range row {
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	termWidth := getTerminalWidth()
	totalPad := (len(headers)-1)*colPadding + indent*2
	available := termWidth - totalPad

	total := 0
	for _, w := range widths {
		total += w
	}

	if total > available {
		for i := range widths {
			widths[i] = max(minColWidth, widths[i]*available/total)
		}
	}

	return widths
}

type rawField struct {
	key   string
	value json.RawMessage
}

func orderedFields(data json.RawMessage) ([]rawField, bool) {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return nil, false
	}

	var fields []rawField
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, false
		}
		key, ok := t.(string)
		if !ok {
			return nil, false
		}

		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, false
		}
		fields = append(fields, rawField{key, value})
	}
	return fields, true
}

func singleNestedObject(data json.RawMessage) (json.RawMessage, bool) {
	fields, ok := orderedFields(data)
	if !ok || len(fields) != 1 {
		return nil, false
	}
	if isJSONObject(fields[0].value) {
		return fields[0].value, true
	}
	return nil, false
}

func printObject(data json.RawMessage, indent int) {
	fields, ok := orderedFields(data)
	if !ok {
		printIndentedJSON(data, indent)
		return
	}

	linePrefix := indentPrefix(indent)

	values := make([]string, len(fields))
	inline := make([]bool, len(fields))
	maxLabelLen := 0
	for i, f := range fields {
		values[i], inline[i] = inlineField(f.value)
		if inline[i] {
			if l := len(formatHeader(f.key)); l > maxLabelLen {
				maxLabelLen = l
			}
		}
	}

	for i, f := range fields {
		label := formatHeader(f.key)
		switch {
		case inline[i]:
			fmt.Printf("%s%-*s  %s\n", linePrefix, maxLabelLen, label, values[i])
		case isJSONObject(f.value):
			fmt.Printf("%s%s\n", linePrefix, label)
			printObject(f.value, indent+1)
		default:
			fmt.Printf("%s%s\n", linePrefix, label)
			printArray(f.value, indent+1)
		}
	}
}

func printArray(data json.RawMessage, indent int) {
	var rows []map[string]interface{}
	if err := json.Unmarshal(data, &rows); err == nil && len(rows) > 0 {
		printTable(rows, indent)
		return
	}
	printIndentedJSON(data, indent)
}

func printIndentedJSON(v interface{}, indent int) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return
	}
	prefix := indentPrefix(indent)
	for _, line := range strings.Split(string(out), "\n") {
		fmt.Printf("%s%s\n", prefix, line)
	}
}

func indentPrefix(indent int) string {
	return strings.Repeat("  ", indent)
}

func isJSONObject(data json.RawMessage) bool {
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func inlineField(data json.RawMessage) (string, bool) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return strings.TrimSpace(string(data)), true
	}
	switch val := v.(type) {
	case map[string]interface{}:
		if len(val) > 0 {
			return "", false
		}
	case []interface{}:
		for _, item := range val {
			if !isScalar(item) {
				return "", false
			}
		}
	}
	if isScalar(v) {
		return formatValue(v), true
	}
	compact, err := json.Marshal(v)
	if err != nil {
		return strings.TrimSpace(string(data)), true
	}
	return string(compact), true
}

func printToon(data json.RawMessage) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return
	}

	out, err := gotoon.Encode(v)
	if err != nil {
		fmt.Println(string(data))
		return
	}

	fmt.Print(out)
}

func printToonPagination(pagination json.RawMessage) {
	if pagination == nil {
		return
	}

	var v interface{}
	if err := json.Unmarshal(pagination, &v); err != nil {
		return
	}

	out, err := gotoon.Encode(v)
	if err != nil {
		return
	}

	fmt.Print(out)
}

func isScalar(v interface{}) bool {
	switch v.(type) {
	case nil, string, float64, bool, json.Number:
		return true
	}
	return false
}

func formatHeader(field string) string {
	return strings.ToUpper(strings.ReplaceAll(field, "_", " "))
}

func formatValue(v interface{}) string {
	if v == nil {
		return "-"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "Yes"
		}
		return "No"
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func getTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 100
	}
	return w
}

func printPaginationSummary(pagination json.RawMessage) {
	if pagination == nil {
		return
	}

	var p struct {
		CurrentPageNumber int `json:"current_page_number"`
		TotalPages        int `json:"total_pages"`
		TotalRecords      int `json:"total_records"`
	}
	if err := json.Unmarshal(pagination, &p); err != nil {
		return
	}

	fmt.Printf("\nPage %d of %d (%d total records)\n", p.CurrentPageNumber, p.TotalPages, p.TotalRecords)
}

func printBreadcrumbs(breadcrumbs []Breadcrumb) {
	if len(breadcrumbs) == 0 {
		return
	}

	fmt.Println()
	for _, b := range breadcrumbs {
		fmt.Printf("  %s: %s\n", b.Label, b.Command)
	}
}
