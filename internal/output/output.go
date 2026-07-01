// Package output renders command results as deterministic JSON (default,
// agent/CI-friendly) or a human table when stdout is an interactive terminal.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"golang.org/x/term"
)

// Format selects how command results are rendered.
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
)

// Resolve maps the --output flag (possibly empty) and TTY-ness of stdout to a
// concrete format: explicit flag wins; otherwise table on a TTY, JSON not.
func Resolve(flag string, isTTY bool) (Format, error) {
	switch flag {
	case "":
		if isTTY {
			return FormatTable, nil
		}
		return FormatJSON, nil
	case string(FormatJSON):
		return FormatJSON, nil
	case string(FormatTable):
		return FormatTable, nil
	default:
		return "", fmt.Errorf("invalid --output %q: must be json or table", flag)
	}
}

// IsTTY reports whether w is an interactive terminal.
func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// RenderJSON writes v as indented JSON followed by a newline.
func RenderJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// RenderTable writes an aligned text table with a header row.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	printRow := func(cells []string) {
		for i, c := range cells {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, c)
		}
		fmt.Fprintln(tw)
	}
	printRow(headers)
	for _, row := range rows {
		printRow(row)
	}
	return tw.Flush()
}
