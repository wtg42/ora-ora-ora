package storage

// FileStorage 以 JSONL 方式將筆記寫入檔案，方便日後以 Bleve 建索引。
// 檔案位置：<baseDir>/YYYY-MM-DD.jsonl 一行一筆。

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"
)

type FileStorage struct {
    baseDir string
}

// NewFileStorage 建立檔案型儲存，若路徑不存在會嘗試建立資料夾。
func NewFileStorage(baseDir string) (*FileStorage, error) {
    if baseDir == "" {
        return nil, errors.New("empty baseDir")
    }
    if err := os.MkdirAll(baseDir, 0o755); err != nil {
        return nil, fmt.Errorf("mkdir %s: %w", baseDir, err)
    }
    return &FileStorage{baseDir: baseDir}, nil
}

// Save 以追加方式寫入 JSONL。
func (f *FileStorage) Save(note Note) error {
    if note.ID == "" {
        return errors.New("note.ID is empty")
    }
    date := note.CreatedAt
    if date.IsZero() {
        date = time.Now()
    }
    fname := fmt.Sprintf("%s.jsonl", date.Format("2006-01-02"))
    path := filepath.Join(f.baseDir, fname)
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return fmt.Errorf("ensure dir: %w", err)
    }
    fp, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
    if err != nil {
        return fmt.Errorf("open %s: %w", path, err)
    }
    defer fp.Close()
    enc, err := json.Marshal(note)
    if err != nil {
        return fmt.Errorf("marshal note: %w", err)
    }
    if _, err := fp.Write(append(enc, '\n')); err != nil {
        return fmt.Errorf("write %s: %w", path, err)
    }
    return nil
}

// List 讀取 baseDir 下所有 .jsonl，逐行解碼為 Note。
func (f *FileStorage) List() ([]Note, error) {
    entries, err := os.ReadDir(f.baseDir)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return nil, nil
        }
        return nil, fmt.Errorf("readdir %s: %w", f.baseDir, err)
    }
    var out []Note
    for _, e := range entries {
        if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
            continue
        }
        path := filepath.Join(f.baseDir, e.Name())
        if err := readJSONL(path, &out); err != nil {
            return nil, err
        }
    }
    sort.Slice(out, func(i, j int) bool {
        return out[i].CreatedAt.After(out[j].CreatedAt)
    })
    return out, nil
}

func readJSONL(path string, out *[]Note) error {
    fp, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open %s: %w", path, err)
    }
    defer fp.Close()
    r := bufio.NewReader(fp)
    for {
        line, err := r.ReadBytes('\n')
        if errors.Is(err, io.EOF) {
            if len(line) == 0 {
                break
            }
            // fallthrough to decode last line without trailing newline
        } else if err != nil {
            return fmt.Errorf("read %s: %w", path, err)
        }
        lineStr := strings.TrimSpace(string(line))
        if lineStr == "" {
            if errors.Is(err, io.EOF) {
                break
            }
            continue
        }
        var n Note
        if derr := json.Unmarshal([]byte(lineStr), &n); derr != nil {
            return fmt.Errorf("decode %s: %w", path, derr)
        }
        *out = append(*out, n)
        if errors.Is(err, io.EOF) {
            break
        }
    }
    return nil
}
