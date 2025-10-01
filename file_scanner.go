package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileItem struct {
	Path     string
	Name     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	Preview  string
	Keep     bool
	Decided  bool
	Skipped  bool
}

func scanDirectory(dir string) ([]FileItem, error) {
	var items []FileItem
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if path == dir {
			return nil
		}
		
		info, err := d.Info()
		if err != nil {
			return err
		}
		
		relPath, _ := filepath.Rel(dir, path)
		if strings.HasPrefix(relPath, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		preview := ""
		if !d.IsDir() && info.Size() < 10240 { // Only preview files < 10KB
			preview = getFilePreview(path)
		}

		item := FileItem{
			Path:    path,
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Preview: preview,
			Keep:    false,
			Decided: false,
			Skipped: false,
		}
		
		items = append(items, item)
		
		if d.IsDir() {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return items, err
}

func getFilePreview(path string) string {
	if !isTextFile(path) {
		return ""
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineCount := 0
	maxLines := 3
	
	// Show more lines for code files
	if isCodeFile(path) {
		maxLines = 15 // More lines for the dedicated code box
	}

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" || isCodeFile(path) {
			lines = append(lines, line)
			lineCount++
		}
	}

	if len(lines) == 0 {
		return ""
	}

	preview := strings.Join(lines, "\n")
	if len(preview) > 800 { // Allow more content for code files
		preview = preview[:797] + "..."
	}

	return preview
}

func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := []string{
		".txt", ".md", ".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h",
		".css", ".html", ".xml", ".json", ".yaml", ".yml", ".toml", ".ini",
		".sh", ".bat", ".ps1", ".sql", ".r", ".php", ".rb", ".scala", ".kt",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	return false
}

func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := []string{
		".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".hpp",
		".rs", ".php", ".rb", ".swift", ".kt", ".scala", ".jsx", ".tsx",
		".vue", ".css", ".scss", ".sass", ".html", ".xml", ".json",
		".yaml", ".yml", ".toml", ".sh", ".bash", ".zsh", ".fish",
		".bat", ".ps1", ".sql", ".r", ".m", ".cs", ".vb", ".pl", ".lua",
	}

	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}

	filename := strings.ToLower(filepath.Base(path))
	if filename == "dockerfile" || filename == "makefile" {
		return true
	}

	return false
}

func getFileIcon(path string, isDir bool) string {
	if isDir {
		return "📁"
	}

	ext := strings.ToLower(filepath.Ext(path))
	
	iconMap := map[string]string{
		// Code files
		".go":     "🐹",
		".js":     "🟨",
		".ts":     "🔷",
		".py":     "🐍",
		".java":   "☕",
		".c":      "🔧",
		".cpp":    "🔧",
		".h":      "📋",
		".rs":     "🦀",
		".php":    "🐘",
		".rb":     "💎",
		".swift":  "🍎",
		".kt":     "🟣",
		".scala":  "🔴",
		
		// Web files
		".html":   "🌐",
		".css":    "🎨",
		".scss":   "🎨",
		".sass":   "🎨",
		".jsx":    "⚛️",
		".tsx":    "⚛️",
		".vue":    "💚",
		
		// Data files
		".json":   "📋",
		".xml":    "📋",
		".yaml":   "📋",
		".yml":    "📋",
		".toml":   "📋",
		".ini":    "⚙️",
		".cfg":    "⚙️",
		".conf":   "⚙️",
		
		// Documents
		".md":     "📝",
		".txt":    "📄",
		".pdf":    "📕",
		".doc":    "📘",
		".docx":   "📘",
		".xls":    "📗",
		".xlsx":   "📗",
		".ppt":    "📙",
		".pptx":   "📙",
		
		// Images
		".jpg":    "🖼️",
		".jpeg":   "🖼️",
		".png":    "🖼️",
		".gif":    "🖼️",
		".svg":    "🎨",
		".ico":    "🖼️",
		".webp":   "🖼️",
		".bmp":    "🖼️",
		
		// Audio
		".mp3":    "🎵",
		".wav":    "🎵",
		".flac":   "🎵",
		".m4a":    "🎵",
		".ogg":    "🎵",
		
		// Video
		".mp4":    "🎬",
		".avi":    "🎬",
		".mkv":    "🎬",
		".mov":    "🎬",
		".wmv":    "🎬",
		".flv":    "🎬",
		".webm":   "🎬",
		
		// Archives
		".zip":    "📦",
		".tar":    "📦",
		".gz":     "📦",
		".rar":    "📦",
		".7z":     "📦",
		".bz2":    "📦",
		".xz":     "📦",
		
		// Executables
		".exe":    "⚡",
		".app":    "📱",
		".deb":    "📦",
		".rpm":    "📦",
		".dmg":    "💿",
		".iso":    "💿",
		
		// System files
		".log":    "📋",
		".tmp":    "🗑️",
		".cache":  "🗑️",
		".bak":    "💾",
		".old":    "💾",
		
		// Shell scripts
		".sh":     "🐚",
		".bash":   "🐚",
		".zsh":    "🐚",
		".fish":   "🐚",
		".bat":    "🖥️",
		".ps1":    "🔷",
		
		// Database
		".db":     "🗄️",
		".sqlite": "🗄️",
		".sql":    "🗄️",
		
		// Git
		".git":    "🔀",
		
		// Docker
		"dockerfile": "🐳",
	}
	
	// Check for special filenames without extensions
	filename := strings.ToLower(filepath.Base(path))
	if filename == "dockerfile" || filename == "makefile" || filename == "readme" {
		if icon, exists := iconMap[filename]; exists {
			return icon
		}
	}
	
	if icon, exists := iconMap[ext]; exists {
		return icon
	}
	
	return "📄" // Default file icon
}
