//go:build ignore

// export_resource_definitions prints the canonical resource definitions (JSON)
// derived from `models.GetDefinitions()`. Use this to produce a JSON file for
// frontend consumption or build-time embedding.
//
// Usage:
//   go run -tags local files/src/scripts/export_resource_definitions/main.go -pretty=true
//   go run -tags local files/src/scripts/export_resource_definitions/main.go -out=resource_definitions.json
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"

    "github.com/fpgschiba/volleygoals/models"
)

func main() {
    pretty := flag.Bool("pretty", true, "Pretty-print JSON output")
    out := flag.String("out", "", "Write JSON to file (if empty, prints to stdout)")
    flag.Parse()

    defs := models.GetDefinitions()
    var b []byte
    var err error
    if *pretty {
        b, err = json.MarshalIndent(defs, "", "  ")
    } else {
        b, err = json.Marshal(defs)
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, "marshal definitions:", err)
        os.Exit(1)
    }
    if *out != "" {
        if err := os.WriteFile(*out, b, 0644); err != nil {
            fmt.Fprintln(os.Stderr, "write file:", err)
            os.Exit(1)
        }
        fmt.Println("Wrote", *out)
        return
    }
    fmt.Println(string(b))
}

