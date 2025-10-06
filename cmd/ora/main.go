// Package main 是 Ora 應用程式的進入點。
// 它使用 Cobra 框架來構建命令列介面 (CLI)。
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wtg42/ora-ora-ora/internal/note"
	"github.com/wtg42/ora-ora-ora/internal/storage"
)

// rootCmd 是整個 Ora 應用程式的基礎命令。
// 所有的子命令都將註冊到此命令下。
var rootCmd = &cobra.Command{
	Use:   "ora",
	Short: "Ora 是一個 AI 快速筆記應用程式",
	Long:  `Ora 是一個用於快速建立和管理筆記的命令列應用程式，旨在與 AI CLI 代理互動。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果沒有給定子命令，則執行此處的預設行為。
		fmt.Println("歡迎使用 Ora！使用 'ora --help' 獲取更多資訊。")
	},
}

// noteCmd 是一個用於管理筆記的子命令。
// 它包含了建立、查看和管理筆記的相關功能。
var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "管理您的筆記",
	Long:  `提供用於建立、查看和管理筆記的命令。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果沒有給定 note 子命令，則顯示 note 命令的幫助資訊。
		cmd.Help()
	},
}

// noteNewCmd 是一個用於建立新筆記的子命令。
// 它會引導使用者輸入筆記標題、內容和可選標籤。
var noteNewCmd = &cobra.Command{
	Use:   "new",
	Short: "建立一個新筆記",
	Long:  `透過提示輸入標題、內容和可選標籤來互動式地建立一個新筆記。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 執行應用程式初始化，獲取配置和資料目錄。
		configDir, dataDir, err := runApp()
		if err != nil {
			log.Fatalf("應用程式錯誤: %v", err)
		}

		fmt.Printf("配置目錄: %s\n", configDir)
		fmt.Printf("資料目錄: %s\n", dataDir)

		// 建立一個讀取器以從標準輸入讀取使用者輸入。
		reader := bufio.NewReader(os.Stdin)

		// 提示使用者輸入筆記標題。
		fmt.Print("輸入筆記標題: ")
		title, _ := reader.ReadString('\n')
		title = strings.TrimSpace(title)

		// 提示使用者輸入筆記內容，直到輸入兩次 Enter 為止。
		fmt.Print("輸入筆記內容 (按兩次 Enter 結束):\n")
		var contentBuilder strings.Builder
		for {
			line, _ := reader.ReadString('\n')
			if strings.TrimSpace(line) == "" {
				break
			}
			contentBuilder.WriteString(line)
		}
		content := strings.TrimSpace(contentBuilder.String())

		fmt.Print("輸入標籤 (逗號分隔，可選): ")
		tagsInput, _ := reader.ReadString('\n')
		tagsInput = strings.TrimSpace(tagsInput)
		var tags []string
		if tagsInput != "" {
			// 將輸入的標籤字串分割並去除空格。
			tags = strings.Split(tagsInput, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}

		// 建立一個新的筆記物件。
		newNote := note.NewNote(title, content, tags)
		fmt.Printf("\n新筆記已建立:\n")
		fmt.Printf("標題: %s\n", newNote.Title)
		fmt.Printf("內容: %s\n", newNote.Content)
		fmt.Printf("標籤: %v\n", newNote.Tags)
		fmt.Printf("建立時間: %s\n", newNote.CreatedAt.Format(time.RFC3339))

		// 儲存新建立的筆記。
		err = storage.SaveNote(newNote)
		if err != nil {
			log.Fatalf("儲存筆記失敗: %v", err)
		}
		fmt.Println("\n筆記已成功建立並儲存！")
	},
}

// runApp 包含應用程式的核心邏輯。
// 它返回配置和資料目錄的路徑，如果獲取失敗則返回錯誤。
func runApp() (string, string, error) {
	// 獲取配置目錄。
	configDir, err := storage.GetConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("獲取配置目錄失敗: %w", err)
	}

	// 獲取資料目錄。
	dataDir, err := storage.GetDataDir()
	if err != nil {
		return "", "", fmt.Errorf("獲取資料目錄失敗: %w", err)
	}

	return configDir, dataDir, nil
}

// init 函數在 main 函數執行前被呼叫，用於初始化 Cobra 命令。
func init() {
	// 將 noteCmd 添加為 rootCmd 的子命令。
	rootCmd.AddCommand(noteCmd)
	// 將 noteNewCmd 添加為 noteCmd 的子命令。
	noteCmd.AddCommand(noteNewCmd)
}

// main 函數是應用程式的入口點。
func main() {
	// 執行 rootCmd，解析命令列參數並執行對應的命令。
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "錯誤: %v\n", err)
		os.Exit(1)
	}
}
