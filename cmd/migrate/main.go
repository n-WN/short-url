package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"short-url/internal/config"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	conn, err := pgx.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// 创建迁移表
	if err := createMigrationTable(context.Background(), conn); err != nil {
		log.Fatalf("Failed to create migration table: %v", err)
	}

	// 运行迁移
	if err := runMigrations(context.Background(), conn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Database migration completed successfully!")
}

func createMigrationTable(ctx context.Context, conn *pgx.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := conn.Exec(ctx, query)
	return err
}

func runMigrations(ctx context.Context, conn *pgx.Conn) error {
	// 获取已应用的迁移
	appliedMigrations, err := getAppliedMigrations(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 获取所有迁移文件
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// 应用未执行的迁移
	for _, file := range migrationFiles {
		version := strings.TrimSuffix(file, ".sql")

		if _, exists := appliedMigrations[version]; exists {
			fmt.Printf("Migration %s already applied, skipping\n", version)
			continue
		}

		fmt.Printf("Applying migration %s...\n", version)

		if err := applyMigration(ctx, conn, file, version); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}
	}

	return nil
}

func getAppliedMigrations(ctx context.Context, conn *pgx.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func getMigrationFiles() ([]string, error) {
	var files []string

	migrationDir := "sql/migrations"
	err := filepath.WalkDir(migrationDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, filepath.Base(path))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 排序确保按顺序执行
	sort.Strings(files)
	return files, nil
}

func applyMigration(ctx context.Context, conn *pgx.Conn, filename, version string) error {
	// 读取迁移文件
	content, err := os.ReadFile(filepath.Join("sql/migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// 开始事务
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 执行迁移SQL
	_, err = tx.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// 记录迁移
	_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// 提交事务
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
