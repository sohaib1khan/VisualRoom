package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirSizes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := make([]byte, 2048)
	if err := os.WriteFile(filepath.Join(sub, "b.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}

	w := New(Options{Workers: 2})
	node, err := w.ScanDir(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if node.Size < 2053 {
		t.Fatalf("expected size >= 2053, got %d", node.Size)
	}
	if len(node.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(node.Children))
	}
	if node.Children[0].Size < node.Children[1].Size {
		t.Fatal("children should be sorted by size descending")
	}
}
