package docs_test

import (
	"strings"
	"testing"

	_ "qr-parking/docs"
	"github.com/swaggo/swag"
)

func TestReadDocSubstitutesHost(t *testing.T) {
	d, err := swag.ReadDoc()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(d, "{{.Host}}") {
		t.Fatalf("ReadDoc returned raw template; host not substituted. First 400 chars:\n%s", d[:min(400, len(d))])
	}
	if !strings.Contains(d, `"host"`) {
		t.Fatal("invalid json")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
