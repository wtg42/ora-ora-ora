package storage

import (
    "context"
    "encoding/json"
    "fmt"
    "bufio"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// jsonlStorage 為 JSONL 檔案後端的最小骨架實作。
// 實作細節（路徑、檔名、錯誤處理）請見下方 TODO。
type jsonlStorage struct {
    baseDir string
}

// Save 將單筆 Note 以 JSONL 方式寫入。
// 建議行為：
// - 目的路徑：<baseDir>/YYYY-MM-DD.jsonl（依 note.CreatedAt）
// - 旗標：os.O_CREATE|os.O_APPEND；自動建立目錄
// - 錯誤處理：將 I/O/序列化錯誤包裝回傳，供上層提示
// TODO: 實作寫入邏輯（建立資料夾、編碼 JSON、逐行寫入）。
func (s *jsonlStorage) Save(note Note) error { // TODO: implement
    if s == nil {
        return fmt.Errorf("jsonl storage is nil")
    }
    if s.baseDir == "" {
        return fmt.Errorf("jsonl storage baseDir is empty")
    }

    // 確保目錄存在（多次呼叫安全）。
    if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
        return fmt.Errorf("create base dir %s: %w", s.baseDir, err)
    }

    // 以 CreatedAt 的 UTC 日期作為檔名（YYYY-MM-DD.jsonl）。
    fname := note.CreatedAt.UTC().Format("2006-01-02") + ".jsonl"
    fpath := filepath.Join(s.baseDir, fname)

    // 以附加模式開檔，若不存在則建立。
    f, err := os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
    if err != nil {
        return fmt.Errorf("open %s for append: %w", fpath, err)
    }
    defer func() { _ = f.Close() }()

    // 單行一筆：序列化後加換行，避免部分寫入需報錯。
    b, err := json.Marshal(note)
    if err != nil {
        return fmt.Errorf("marshal note %s: %w", note.ID, err)
    }
    if _, err := f.Write(append(b, '\n')) ; err != nil {
        return fmt.Errorf("write %s: %w", fpath, err)
    }
    return nil
}

// List 讀取所有 JSONL 檔案並反序列化為 Note 列表。
// 建議行為：
// - 掃描 baseDir 下所有 *.jsonl，依時間或寫入順序確保穩定性
// - 單行壞資料：建議直接回傳錯誤（或以選項決定是否跳過）
// - 回傳完整 Note 列表
// TODO: 實作讀取與反序列化邏輯（遍歷檔案、逐行 decode）。
func (s *jsonlStorage) List() ([]Note, error) { // TODO: implement
    if s == nil {
        return nil, fmt.Errorf("jsonl storage is nil")
    }
    if s.baseDir == "" {
        return nil, fmt.Errorf("jsonl storage baseDir is empty")
    }

    // 若目錄不存在，視為無資料，回傳空列表。
    if _, err := os.Stat(s.baseDir); err != nil {
        if os.IsNotExist(err) {
            return []Note{}, nil
        }
        return nil, fmt.Errorf("stat base dir %s: %w", s.baseDir, err)
    }

    entries, err := os.ReadDir(s.baseDir)
    if err != nil {
        return nil, fmt.Errorf("read dir %s: %w", s.baseDir, err)
    }

    // 收集所有 .jsonl 檔並依名稱排序，確保跨檔案的穩定性。
    var files []string
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        if strings.EqualFold(filepath.Ext(e.Name()), ".jsonl") {
            files = append(files, filepath.Join(s.baseDir, e.Name()))
        }
    }
    sort.Strings(files)

    var out []Note
    for _, fp := range files {
        f, err := os.Open(fp)
        if err != nil {
            return nil, fmt.Errorf("open %s: %w", fp, err)
        }
        scanner := bufio.NewScanner(f)
        lineNo := 0
        for scanner.Scan() {
            lineNo++
            raw := strings.TrimSpace(scanner.Text())
            if raw == "" {
                continue // 忽略空行
            }
            var n Note
            if err := json.Unmarshal([]byte(raw), &n); err != nil {
                _ = f.Close()
                return nil, fmt.Errorf("decode %s line %d: %w", fp, lineNo, err)
            }
            out = append(out, n)
        }
        if err := scanner.Err(); err != nil {
            _ = f.Close()
            return nil, fmt.Errorf("scan %s: %w", fp, err)
        }
        _ = f.Close()
    }
    return out, nil
}

// 可選：保留 context 擴充點，以便未來支援取消/逾時或背景作業。
// 目前未使用，僅示意未來擴充介面位置。
var _ = context.TODO
