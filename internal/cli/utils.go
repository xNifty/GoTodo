package cli

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

func ReturnHeader(headers []string) string {
	fmt.Println("headers: ", headers)
	if len(headers) == 0 {
		return "There was no string provided."
	}

	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 0, 4, ' ', 0)

	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	seperator := make([]string, len(headers))

	for i := range seperator {
		seperator[i] = strings.Repeat("-", len(headers[i]))
	}

	writer.Flush()
	fmt.Println("buf: ", buf.String())
	return buf.String()
}

func ReturnUnderline(headers []string) string {
	underlines := make([]string, len(headers))
	for i, header := range headers {
		underlines[i] = strings.Repeat("-", len(header))
	}

	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 0, 4, ' ', 0)
	_, err := fmt.Fprintf(writer, strings.Join(underlines, "\t")+"\n")

	if err != nil {
		return "Error in the ReturnUnderline function." + err.Error()
	}

	writer.Flush()
	return buf.String()
}
