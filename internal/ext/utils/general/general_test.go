package general_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m0090-dev/eec/internal/ext/utils/general"
)

// =============================================================
// file.go — 純粋関数（パス操作）のテスト
// FS を直接触る GetFilesWithExtension と FileExists だけ
// t.TempDir() を使い、テスト後は自動削除される
// =============================================================

func TestFileExists_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exist.txt")
	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if !general.FileExists(path) {
		t.Errorf("FileExists(%q) = false, want true", path)
	}
}

func TestFileExists_NonExistentFile(t *testing.T) {
	if general.FileExists("/absolutely/nonexistent/path/file.txt") {
		t.Error("FileExists(nonexistent) = true, want false")
	}
}

func TestFileExt(t *testing.T) {
	cases := []struct{ in, want string }{
		{"foo.txt", ".txt"},
		{"foo.tar.gz", ".gz"},
		{"noext", ""},
		{"/path/to/file.toml", ".toml"},
		{"file.YAML", ".YAML"},
	}
	for _, c := range cases {
		if got := general.FileExt(c.in); got != c.want {
			t.Errorf("FileExt(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFileBase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"/path/to/file.txt", "file.txt"},
		{"just.txt", "just.txt"},
	}
	for _, c := range cases {
		if got := general.FileBase(c.in); got != c.want {
			t.Errorf("FileBase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAddExtension(t *testing.T) {
	cases := []struct {
		filename, ext, want string
	}{
		{"foo", ".txt", "foo.txt"},
		{"foo.txt", ".txt", "foo.txt"},     // すでに付いている
		{"foo.TXT", ".txt", "foo.TXT"},     // 大文字小文字違いは一致とみなす
		{"foo", "", "foo"},                 // ext 空
		{"foo.bar", ".txt", "foo.bar.txt"}, // 別拡張子は追記
	}
	for _, c := range cases {
		got := general.AddExtension(c.filename, c.ext)
		if got != c.want {
			t.Errorf("AddExtension(%q, %q) = %q, want %q", c.filename, c.ext, got, c.want)
		}
	}
}

func TestExtractFileNameSafe(t *testing.T) {
	cases := []struct {
		input, start, ext, want string
	}{
		{"path/to/config.toml rest", "config", ".toml", "config.toml"},
		{"nothing here", "config", ".toml", ""},
		{"file.env and other", "file", ".env", "file.env"},
	}
	for _, c := range cases {
		got := general.ExtractFileNameSafe(c.input, c.start, c.ext)
		if got != c.want {
			t.Errorf("ExtractFileNameSafe(%q, %q, %q) = %q, want %q",
				c.input, c.start, c.ext, got, c.want)
		}
	}
}

func TestRemoveExtension(t *testing.T) {
	cases := []struct{ in, want string }{
		{"foo.txt", "foo"},
		{"noext", "noext"},
		{"foo.tar.gz", "foo.tar"},
		{"eec.exe", "eec"},
	}
	for _, c := range cases {
		if got := general.RemoveExtension(c.in); got != c.want {
			t.Errorf("RemoveExtension(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRemoveExtension_WithDir(t *testing.T) {
	got := general.RemoveExtension("dir/sub/file.toml")
	if !strings.HasSuffix(got, "file") {
		t.Errorf("RemoveExtension with dir path: got %q, want suffix 'file'", got)
	}
}

func TestRemoveExtensions(t *testing.T) {
	in := []string{"a.txt", "b.go", "noext"}
	want := []string{"a", "b", "noext"}
	got := general.RemoveExtensions(in)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("RemoveExtensions[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBaseSlice(t *testing.T) {
	in := []string{"/a/b/foo.txt", "/x/y/bar.go"}
	want := []string{"foo.txt", "bar.go"}
	got := general.BaseSlice(in)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("BaseSlice[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGetFilesWithExtension_ReturnsOnlyMatchingExt(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.tag", "b.tag", "c.txt"} {
		f, _ := os.Create(filepath.Join(dir, name))
		f.Close()
	}
	files, err := general.GetFilesWithExtension(dir, ".tag")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("got %d files, want 2", len(files))
	}
	for _, f := range files {
		if filepath.Ext(f) != ".tag" {
			t.Errorf("unexpected extension in result: %s", f)
		}
	}
}

func TestGetFilesWithExtension_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := general.GetFilesWithExtension(dir, ".tag")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestGetFilesWithExtension_NonExistentDir_ReturnsError(t *testing.T) {
	_, err := general.GetFilesWithExtension("/nonexistent/path_xyz", ".tag")
	if err == nil {
		t.Error("expected error for nonexistent dir, got nil")
	}
}

func TestNormalizeSourcePath_Empty(t *testing.T) {
	if got := general.NormalizeSourcePath(""); got != "" {
		t.Errorf("NormalizeSourcePath(\"\") = %q, want \"\"", got)
	}
}

func TestNormalizeSourcePath_RelativeBecomesAbsolute(t *testing.T) {
	got := general.NormalizeSourcePath("somefile.toml")
	if !filepath.IsAbs(got) {
		t.Errorf("NormalizeSourcePath returned non-absolute path: %q", got)
	}
}

func TestNormalizeSourcePath_AlreadyAbsolute(t *testing.T) {
	abs := filepath.FromSlash("/some/absolute/path.toml")
	expected, _ := filepath.Abs(abs)
	got := general.NormalizeSourcePath(abs)
	if got != expected {
		t.Errorf("NormalizeSourcePath(%q) = %q, want %q", abs, got, expected)
	}
}

// =============================================================
// print.go — 標準出力へ書くだけの関数。panic しないことを確認。
// =============================================================

func TestPrintBlock_NormalCase_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintBlock panicked: %v", r)
		}
	}()
	general.PrintBlock("Test Title", map[string]interface{}{
		"key1":      "value1",
		"key2":      42,
		"multiline": "line1\nline2\nline3",
	})
}

func TestPrintBlock_EmptyMap_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintBlock with empty map panicked: %v", r)
		}
	}()
	general.PrintBlock("Empty", map[string]interface{}{})
}

func TestPrintBlock_EmptyTitle_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintBlock with empty title panicked: %v", r)
		}
	}()
	general.PrintBlock("", map[string]interface{}{"k": "v"})
}

// =============================================================
// string.go — バイトバッファを使ったラウンドトリップテスト
// (os 一切不使用)
// =============================================================

func TestWriteReadString_RoundTrip(t *testing.T) {
	cases := []string{
		"hello",
		"",
		"日本語",
		"a very long string with spaces & symbols !@#$%",
	}
	for _, s := range cases {
		var buf bytes.Buffer
		if err := general.WriteString(&buf, s); err != nil {
			t.Fatalf("WriteString(%q): %v", s, err)
		}
		r := bytes.NewReader(buf.Bytes())
		got, err := general.ReadString(r)
		if err != nil {
			t.Fatalf("ReadString(%q): %v", s, err)
		}
		if got != s {
			t.Errorf("round-trip: got %q, want %q", got, s)
		}
	}
}

func TestWriteReadStringSlice_RoundTrip(t *testing.T) {
	cases := [][]string{
		{"a", "b", "c"},
		{},
		{"single"},
		{"", "empty", ""},
	}
	for _, ss := range cases {
		var buf bytes.Buffer
		if err := general.WriteStringSlice(&buf, ss); err != nil {
			t.Fatalf("WriteStringSlice(%v): %v", ss, err)
		}
		r := bytes.NewReader(buf.Bytes())
		got, err := general.ReadStringSlice(r)
		if err != nil {
			t.Fatalf("ReadStringSlice(%v): %v", ss, err)
		}
		if len(got) != len(ss) {
			t.Errorf("length mismatch: got %d, want %d", len(got), len(ss))
			continue
		}
		for i := range ss {
			if got[i] != ss[i] {
				t.Errorf("[%d] got %q, want %q", i, got[i], ss[i])
			}
		}
	}
}

func TestReadString_EmptyBuffer_ReturnsError(t *testing.T) {
	r := bytes.NewReader([]byte{})
	_, err := general.ReadString(r)
	if err == nil {
		t.Error("expected error reading from empty buffer, got nil")
	}
}

func TestReadString_TruncatedBuffer_ReturnsError(t *testing.T) {
	// int32 の長さプレフィックスには 4 bytes 必要。2 bytes しかない。
	r := bytes.NewReader([]byte{0x05, 0x00})
	_, err := general.ReadString(r)
	if err == nil {
		t.Error("expected error reading from truncated buffer, got nil")
	}
}

func TestReadStringSlice_EmptyBuffer_ReturnsError(t *testing.T) {
	r := bytes.NewReader([]byte{})
	_, err := general.ReadStringSlice(r)
	if err == nil {
		t.Error("expected error reading from empty buffer, got nil")
	}
}

// WriteString → ReadString の独立性確認：複数回書いても正しく読める
func TestWriteReadString_MultipleWrites(t *testing.T) {
	strs := []string{"first", "second", "third"}
	var buf bytes.Buffer
	for _, s := range strs {
		if err := general.WriteString(&buf, s); err != nil {
			t.Fatalf("WriteString: %v", err)
		}
	}
	r := bytes.NewReader(buf.Bytes())
	for _, want := range strs {
		got, err := general.ReadString(r)
		if err != nil {
			t.Fatalf("ReadString: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
