package tracking

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Record stores a command's token savings to the SQLite database.
func Record(command string, inputTokens, outputTokens int) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		`INSERT INTO commands (ts, command, input_tokens, output_tokens) VALUES (?, ?, ?, ?)`,
		time.Now().Unix(), command, inputTokens, outputTokens,
	)
	return err
}

// CountTokens returns a whitespace-delimited token count (fast approximation).
func CountTokens(text string) int {
	return len(strings.Fields(text))
}

// CommandStat holds aggregated stats for a single command type.
type CommandStat struct {
	Command      string
	Calls        int
	InputTokens  int
	OutputTokens int
	SavingsPct   float64
}

// Gain returns per-command aggregate savings.
func Gain() ([]CommandStat, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT command,
		       COUNT(*) AS calls,
		       SUM(input_tokens) AS total_in,
		       SUM(output_tokens) AS total_out
		FROM commands
		GROUP BY command
		ORDER BY SUM(input_tokens) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []CommandStat
	for rows.Next() {
		var s CommandStat
		if err := rows.Scan(&s.Command, &s.Calls, &s.InputTokens, &s.OutputTokens); err != nil {
			continue
		}
		if s.InputTokens > 0 {
			s.SavingsPct = (1 - float64(s.OutputTokens)/float64(s.InputTokens)) * 100
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func openDB() (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("tracking: create dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("tracking: open db: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS commands (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			ts            INTEGER NOT NULL,
			command       TEXT    NOT NULL,
			input_tokens  INTEGER NOT NULL DEFAULT 0,
			output_tokens INTEGER NOT NULL DEFAULT 0
		)
	`)
	return err
}

func dbPath() (string, error) {
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, "ptk", "tracking.db"), nil
	}
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "ptk", "tracking.db"), nil
}
