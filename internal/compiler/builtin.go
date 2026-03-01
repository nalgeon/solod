package compiler

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed builtin/solod.h builtin/solod.c
var builtinFS embed.FS

func writeBuiltin(outDir string) error {
	for _, name := range []string{"solod.h", "solod.c"} {
		data, err := builtinFS.ReadFile("builtin/" + name)
		if err != nil {
			return fmt.Errorf("read embedded builtin file %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(outDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write builtin file %s: %w", name, err)
		}
	}
	return nil
}
