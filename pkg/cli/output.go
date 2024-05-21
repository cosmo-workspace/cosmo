package cli

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/cli-runtime/pkg/printers"
)

func OutputTable(output io.Writer, headers []string, data [][]string) {
	w := printers.GetNewTabWriter(output)
	defer w.Flush()

	fmt.Fprintf(w, "%s\n", strings.Join(headers, "\t"))
	for _, v := range data {
		fmt.Fprintf(w, "%s\n", strings.Join(v, "\t"))
	}
}
