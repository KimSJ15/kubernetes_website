// +build ignore

// mkstdlib generates the zstdlib.go file, containing the Go standard
// library API symbols. It's baked into the binary to avoid scanning
// GOPATH in the common case.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func mustOpen(name string) io.Reader {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func api(base string) string {
	return filepath.Join(os.Getenv("GOROOT"), "api", base)
}

var sym = regexp.MustCompile(`^pkg (\S+).*?, (?:var|func|type|const) ([A-Z]\w*)`)

func main() {
	var buf bytes.Buffer
	outf := func(format string, args ...interface{}) {
		fmt.Fprintf(&buf, format, args...)
	}
	outf("// AUTO-GENERATED BY mkstdlib.go\n\n")
	outf("package imports\n")
	outf("var stdlib = map[string]string{\n")
	f := io.MultiReader(
		mustOpen(api("go1.txt")),
		mustOpen(api("go1.1.txt")),
		mustOpen(api("go1.2.txt")),
	)
	sc := bufio.NewScanner(f)
	fullImport := map[string]string{} // "zip.NewReader" => "archive/zip"
	ambiguous := map[string]bool{}
	var keys []string
	for sc.Scan() {
		l := sc.Text()
		has := func(v string) bool { return strings.Contains(l, v) }
		if has("struct, ") || has("interface, ") || has(", method (") {
			continue
		}
		if m := sym.FindStringSubmatch(l); m != nil {
			full := m[1]
			key := path.Base(full) + "." + m[2]
			if exist, ok := fullImport[key]; ok {
				if exist != full {
					ambiguous[key] = true
				}
			} else {
				fullImport[key] = full
				keys = append(keys, key)
			}
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if ambiguous[key] {
			outf("\t// %q is ambiguous\n", key)
		} else {
			outf("\t%q: %q,\n", key, fullImport[key])
		}
	}
	outf("}\n")
	fmtbuf, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(fmtbuf)
}
