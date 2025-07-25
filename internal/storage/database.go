package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ProxyWoman/internal/features"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// NewDatabase 创建新的数据库连接
func NewDatabase() (*Database, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	// 创建应用数据目录
	appDataDir := filepath.Join(homeDir, ".proxywoman")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app data directory: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(appDataDir, "proxywoman.db")

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	database := &Database{db: db}

	// 初始化数据库表
	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return database, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// initTables 初始化数据库表
func (d *Database) initTables() error {
	// 创建断点规则表
	breakpointTableSQL := `
	CREATE TABLE IF NOT EXISTS breakpoint_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url_pattern TEXT NOT NULL,
		method TEXT NOT NULL DEFAULT '*',
		enabled BOOLEAN NOT NULL DEFAULT 1,
		is_regex BOOLEAN NOT NULL DEFAULT 0,
		break_on_request BOOLEAN NOT NULL DEFAULT 1,
		break_on_response BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(breakpointTableSQL); err != nil {
		return fmt.Errorf("failed to create breakpoint_rules table: %v", err)
	}

	// 创建脚本表
	scriptTableSQL := `
	CREATE TABLE IF NOT EXISTS scripts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		content TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		type TEXT NOT NULL DEFAULT 'both',
		description TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(scriptTableSQL); err != nil {
		return fmt.Errorf("failed to create scripts table: %v", err)
	}

	return nil
}

// SaveBreakpointRule 保存断点规则
func (d *Database) SaveBreakpointRule(rule *features.BreakpointRule) error {
	query := `
	INSERT OR REPLACE INTO breakpoint_rules 
	(id, name, url_pattern, method, enabled, is_regex, break_on_request, break_on_response, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := d.db.Exec(query,
		rule.ID,
		rule.Name,
		rule.URLPattern,
		rule.Method,
		rule.Enabled,
		rule.IsRegex,
		rule.BreakOnRequest,
		rule.BreakOnResponse,
	)

	return err
}

// GetBreakpointRules 获取所有断点规则
func (d *Database) GetBreakpointRules() ([]*features.BreakpointRule, error) {
	query := `
	SELECT id, name, url_pattern, method, enabled, is_regex, break_on_request, break_on_response, created_at, updated_at
	FROM breakpoint_rules
	ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*features.BreakpointRule
	for rows.Next() {
		rule := &features.BreakpointRule{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.URLPattern,
			&rule.Method,
			&rule.Enabled,
			&rule.IsRegex,
			&rule.BreakOnRequest,
			&rule.BreakOnResponse,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// DeleteBreakpointRule 删除断点规则
func (d *Database) DeleteBreakpointRule(id string) error {
	query := `DELETE FROM breakpoint_rules WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

// SaveScript 保存脚本
func (d *Database) SaveScript(script *features.Script) error {
	query := `
	INSERT OR REPLACE INTO scripts 
	(id, name, content, enabled, type, description, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := d.db.Exec(query,
		script.ID,
		script.Name,
		script.Content,
		script.Enabled,
		script.Type,
		script.Description,
	)

	return err
}

// GetScripts 获取所有脚本
func (d *Database) GetScripts() ([]*features.Script, error) {
	query := `
	SELECT id, name, content, enabled, type, description, created_at, updated_at
	FROM scripts
	ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scripts []*features.Script
	for rows.Next() {
		script := &features.Script{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&script.ID,
			&script.Name,
			&script.Content,
			&script.Enabled,
			&script.Type,
			&script.Description,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析时间
		if script.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt); err != nil {
			script.CreatedAt = time.Now()
		}
		if script.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAt); err != nil {
			script.UpdatedAt = time.Now()
		}

		scripts = append(scripts, script)
	}

	return scripts, nil
}

// DeleteScript 删除脚本
func (d *Database) DeleteScript(id string) error {
	query := `DELETE FROM scripts WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

// UpdateBreakpointRuleStatus 更新断点规则状态
func (d *Database) UpdateBreakpointRuleStatus(id string, enabled bool) error {
	query := `UPDATE breakpoint_rules SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.db.Exec(query, enabled, id)
	return err
}

// UpdateScriptStatus 更新脚本状态
func (d *Database) UpdateScriptStatus(id string, enabled bool) error {
	query := `UPDATE scripts SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.db.Exec(query, enabled, id)
	return err
}
