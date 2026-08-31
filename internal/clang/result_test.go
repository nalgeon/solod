package clang

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestMultiReturnPairs(t *testing.T) {
	// Check that multiReturnPairs lists exactly the (T1, T2) result types of builtin.h.
	path := filepath.Join("..", "compiler", "builtin", "builtin.h")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	re := regexp.MustCompile(`}\s*so_R_(\w+);`)
	want := make(map[string]bool)
	for _, match := range re.FindAllStringSubmatch(string(src), -1) {
		if strings.HasSuffix(match[1], "_err") {
			continue
		}
		want[match[1]] = true
	}
	if len(want) == 0 {
		t.Fatalf("no (T1, T2) result types found in %s", path)
	}

	for pair := range want {
		if !multiReturnPairs[pair] {
			t.Errorf("so_R_%s is in builtin.h, but not in multiReturnPairs", pair)
		}
	}
	for pair := range multiReturnPairs {
		if !want[pair] {
			t.Errorf("so_R_%s is in multiReturnPairs, but not in builtin.h", pair)
		}
	}
}
