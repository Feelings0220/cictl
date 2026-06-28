// Package output renders structured values for humans and agents.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
	FormatMD    Format = "md"
)

func Detect(stdoutIsTTY bool, override string) Format {
	if override != "" {
		return Format(override)
	}
	if stdoutIsTTY {
		return FormatTable
	}
	return FormatJSON
}

func Render(w io.Writer, f Format, value any) error {
	switch f {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(value)
	case FormatTable:
		return renderTable(w, value)
	case FormatMD:
		return renderMD(w, value)
	}
	return fmt.Errorf("unsupported format: %s", f)
}

func rowsOf(value any) ([]map[string]any, error) {
	rows, ok := value.([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("value must be []map[string]any for table/md")
	}
	return rows, nil
}

func columns(rows []map[string]any) []string {
	seen := map[string]struct{}{}
	for _, r := range rows {
		for k := range r {
			seen[k] = struct{}{}
		}
	}
	cols := make([]string, 0, len(seen))
	for k := range seen {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}

func renderTable(w io.Writer, value any) error {
	rows, err := rowsOf(value)
	if err != nil {
		return err
	}
	cols := columns(rows)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, r := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fmt.Sprintf("%v", r[c])
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}

func renderMD(w io.Writer, value any) error {
	rows, err := rowsOf(value)
	if err != nil {
		return err
	}
	cols := columns(rows)
	fmt.Fprintln(w, "| "+strings.Join(cols, " | ")+" |")
	sep := make([]string, len(cols))
	for i := range cols {
		sep[i] = "---"
	}
	fmt.Fprintln(w, "|"+strings.Join(sep, "|")+"|")
	for _, r := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fmt.Sprintf("%v", r[c])
		}
		fmt.Fprintln(w, "| "+strings.Join(cells, " | ")+" |")
	}
	return nil
}
