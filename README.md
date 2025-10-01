# Dinder

> Swipe through your files and folders

A clean TUI file manager that lets you review files one by one, deciding whether to keep or delete them.

https://github.com/user-attachments/assets/04fe4525-6ee8-42a5-ae78-62d53c8be8a9

## Usage

```bash
go run .
```

## Controls

- `→` / `l` / `y` - Keep file
- `←` / `h` / `n` - Delete file
- `s` - Skip file (review later)
- `u` - Undo last decision
- `q` - Quit
- `y` - Confirm deletion
- `n` - Cancel deletion

## Features

- Scans current directory
- One-by-one file review with preview
- File metadata (size, modification date)
- Text file preview (first 3 lines)
- Skip files for later review
- Undo functionality
- Confirmation before deletion
- Progress tracking and completion stats
- Clean TUI with spinners and status indicators
