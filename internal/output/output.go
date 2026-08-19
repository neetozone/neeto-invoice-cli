package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

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
// Ordered for NeetoInvoice payloads: identity first, then who the row is
// about, then the time-tracking and PTO figures, then the surrounding
// client/project/task names, then money and dates.
var priorityFields = []string{
	"sid", "id", "identifier", "name", "title",
	"user_name", "user_email", "email",
	"recorded_on", "date", "hours", "pto_earned", "pto_used", "net_pto",
	"task_name", "project_name", "client_name",
	"status", "role", "kind", "type",
	"number", "total", "amount", "currency", "hourly_rate",
	"issue_date", "due_date", "total_hours", "working_days", "time_zone",
}

const (
	maxTableColumns = 7
	ellipsis        = "..."
	minContentWidth = 10
	minColWidth     = minContentWidth + len(ellipsis)
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
	if msg, ok := thumbsUpNotice(data); ok {
		fmt.Println(msg)
		return
	}

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

func thumbsUpNotice(data json.RawMessage) (string, bool) {
	var obj struct {
		Notice     string `json:"notice"`
		NoticeCode string `json:"notice_code"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", false
	}
	if obj.NoticeCode != "thumbs_up" {
		return "", false
	}
	if obj.Notice != "" {
		return obj.Notice, true
	}
	return "success", true
}

func printTable(rows []map[string]interface{}, indent int) {
	prefix := indentPrefix(indent)

	cols := pickColumns(rows)
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
		cell := truncate(h, widths[i])
		if i < len(headers)-1 {
			cell = padRight(cell, widths[i])
		}
		fmt.Print(cell)
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
			cell := truncate(val, widths[i])
			if i < len(row)-1 {
				cell = padRight(cell, widths[i])
			}
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

func pickColumns(rows []map[string]interface{}) []string {
	scalars := map[string]bool{}
	for k, v := range rows[0] {
		if isScalar(v) {
			scalars[k] = true
		}
	}

	urlFields := map[string]bool{}
	var urlCols []string
	for _, row := range rows {
		for k, v := range row {
			if !scalars[k] || urlFields[k] {
				continue
			}
			if s, ok := v.(string); ok && isURL(s) {
				urlFields[k] = true
				urlCols = append(urlCols, k)
			}
		}
	}
	sort.Strings(urlCols)

	budget := max(1, maxTableColumns-len(urlCols))

	var cols []string
	used := map[string]bool{}

	for _, f := range priorityFields {
		if scalars[f] && !used[f] && !urlFields[f] && len(cols) < budget {
			cols = append(cols, f)
			used[f] = true
		}
	}

	var remaining []string
	for k := range scalars {
		if !used[k] && !urlFields[k] {
			remaining = append(remaining, k)
		}
	}
	sort.Strings(remaining)
	for _, k := range remaining {
		if len(cols) >= budget {
			break
		}
		cols = append(cols, k)
	}

	return append(cols, urlCols...)
}

func calculateWidths(headers []string, grid [][]string, indent int) []int {
	widths := make([]int, len(headers))
	protected := make([]bool, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range grid {
		for i, val := range row {
			if w := displayWidth(val); w > widths[i] {
				widths[i] = w
			}
			if isURL(val) {
				protected[i] = true
			}
		}
	}

	termWidth := getTerminalWidth()
	totalPad := (len(headers)-1)*colPadding + indent*2
	available := termWidth - totalPad

	total, flexible := 0, 0
	for i, w := range widths {
		total += w
		if !protected[i] {
			flexible += w
		}
	}

	if total <= available || flexible == 0 {
		return widths
	}

	budget := max(0, available-(total-flexible))
	for i := range widths {
		if !protected[i] {
			widths[i] = min(widths[i], max(minColWidth, widths[i]*budget/flexible))
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

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func displayWidth(s string) int {
	return utf8.RuneCountInString(s)
}

func padRight(s string, width int) string {
	if gap := width - displayWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

func truncate(s string, maxLen int) string {
	if maxLen < 0 {
		maxLen = 0
	}
	if displayWidth(s) <= maxLen || isURL(s) {
		return s
	}
	runes := []rune(s)
	if maxLen <= len(ellipsis) {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-len(ellipsis)]) + ellipsis
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
