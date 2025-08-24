package storage

import (
    "errors"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

// TestJSONLStorage_SaveAndList_Basic 檢驗最小正向流程：Save 兩筆、List 回傳兩筆且順序穩定。
func TestJSONLStorage_SaveAndList_Basic(t *testing.T) {
    t.Parallel()
    tmp := t.TempDir()
    s := NewJSONL(tmp)

    n1 := Note{ID: "n1", Content: "hello", Tags: []string{"a"}, CreatedAt: mustTime("2024-08-24T10:00:00Z"), UpdatedAt: mustTime("2024-08-24T10:00:00Z")}
    n2 := Note{ID: "n2", Content: "world", Tags: []string{"b"}, CreatedAt: mustTime("2024-08-24T12:00:00Z"), UpdatedAt: mustTime("2024-08-24T12:00:00Z")}

    if err := s.Save(n1); err != nil {
        t.Fatalf("Save n1: %v", err)
    }
    if err := s.Save(n2); err != nil {
        t.Fatalf("Save n2: %v", err)
    }

    got, err := s.List()
    if err != nil {
        t.Fatalf("List: %v", err)
    }
    if len(got) != 2 {
        t.Fatalf("want 2 notes, got %d", len(got))
    }
    if got[0].ID != "n1" || got[1].ID != "n2" {
        t.Fatalf("want order [n1,n2], got [%s,%s]", got[0].ID, got[1].ID)
    }
}

// TestJSONLStorage_AutoCreateDir 若 baseDir 不存在，Save 應能自動建立目錄。
func TestJSONLStorage_AutoCreateDir(t *testing.T) {
    t.Parallel()
    tmp := t.TempDir()
    base := filepath.Join(tmp, "notes") // 尚未建立
    s := NewJSONL(base)

    n := Note{ID: "n1", Content: "auto", CreatedAt: mustTime("2024-08-24T10:00:00Z"), UpdatedAt: mustTime("2024-08-24T10:00:00Z")}
    if err := s.Save(n); err != nil {
        t.Fatalf("Save: %v", err)
    }

    // 確認有產生當日檔案（格式 YYYY-MM-DD.jsonl）
    wantFile := filepath.Join(base, "2024-08-24.jsonl")
    if _, err := os.Stat(wantFile); err != nil {
        t.Fatalf("want created file %s, stat err: %v", wantFile, err)
    }
}

// TestJSONLStorage_ListMalformedLine 驗證讀取壞行時回傳錯誤（而非悄悄略過）。
func TestJSONLStorage_ListMalformedLine(t *testing.T) {
    t.Parallel()
    tmp := t.TempDir()
    base := filepath.Join(tmp, "notes")
    if err := os.MkdirAll(base, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    // 手動寫入一個壞行
    bad := filepath.Join(base, "2024-08-24.jsonl")
    if err := os.WriteFile(bad, []byte("{not-json}\n"), 0o644); err != nil {
        t.Fatalf("write bad file: %v", err)
    }

    s := NewJSONL(base)
    _, err := s.List()
    if err == nil {
        t.Fatalf("want error due to malformed line, got nil")
    }
}

// TestJSONLStorage_SavePermissionError 模擬唯讀目錄下寫入應回錯。
func TestJSONLStorage_SavePermissionError(t *testing.T) {
    t.Parallel()
    tmp := t.TempDir()
    base := filepath.Join(tmp, "ro")
    if err := os.MkdirAll(base, 0o555); err != nil { // 唯讀
        t.Fatalf("mkdir: %v", err)
    }
    t.Cleanup(func() { _ = os.Chmod(base, 0o755) })

    s := NewJSONL(base)
    n := Note{ID: "n1", Content: "ro", CreatedAt: mustTime("2024-08-24T10:00:00Z"), UpdatedAt: mustTime("2024-08-24T10:00:00Z")}
    err := s.Save(n)
    if err == nil {
        t.Fatalf("want permission error, got nil")
    }
    // 不嚴格比對錯誤型態，但訊息應包含 permission 或 read-only 類型字樣。
    msg := strings.ToLower(err.Error())
    if !strings.Contains(msg, "permission") && !strings.Contains(msg, "read-only") {
        t.Fatalf("want permission-related error, got: %v", err)
    }
}

// mustTime 幫手：解析 RFC3339，解析失敗直接 panic（測試限定）。
func mustTime(s string) time.Time {
    tm, err := time.Parse(time.RFC3339, s)
    if err != nil {
        panic(err)
    }
    return tm
}

// 防止未使用匯入報錯（在尚未完成實作時）。
var _ = errors.New

