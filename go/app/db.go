package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
)

func connectDB(c echo.Context) *sql.DB {
	// dbOpen
	db, err := sql.Open("sqlite3", "mercari.sqlite3")
	if err != nil {
		c.Logger().Fatalf("DB接続エラー: %v", err)
	}
	return db
}

func searchCategoryId(db *sql.DB, category string) (int64, error) {
	idStmt, err := db.Prepare("SELECT category_id FROM categories WHERE category_name = ?")
	if err != nil {
		return -1, fmt.Errorf("ステートメント生成エラー: %v", err)
	}
	// stmtClose
	defer func(data *sql.Stmt) {
		_ = data.Close()
	}(idStmt)
	r, err := idStmt.Query(category)
	defer func(r *sql.Rows) {
		_ = r.Close()
	}(r)
	var id int
	r.Next()
	if err := r.Scan(&id); err != nil {
		return -1, fmt.Errorf("既存ID検索エラー: %v", err)
	}
	return int64(id), err
}

func addCategoryId(db *sql.DB, category string) (int64, error) {
	// ステートメントを生成
	categorystmt, err := db.Prepare("INSERT INTO categories (category_name) VALUES (?)")
	if err != nil {
		return -1, fmt.Errorf("ステートメント生成エラー: %v", err)
	}
	// stmtClose
	defer func(data *sql.Stmt) {
		_ = data.Close()
	}(categorystmt)
	fmt.Print(category)
	// ステートメントを用いて書き込み
	result, err := categorystmt.Exec(category)
	if err != nil {
		return -1, fmt.Errorf("新規ID書き込みエラー: %v", err)
	}
	createdId, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("IDの取得エラー: %v", err)
	}
	return createdId, err
}

func checkCategoryId(db *sql.DB, category string) (int64, error) {
	// ステートメントを生成
	stmt, err := db.Prepare("SELECT EXISTS (SELECT * FROM categories WHERE category_name = ?)")
	if err != nil {
		return -1, fmt.Errorf("ステートメント生成エラー: %v", err)
	}
	// DBから取得
	rows, err := stmt.Query(category)
	// stmtClose
	func(data *sql.Stmt) {
		_ = data.Close()
	}(stmt)
	var check bool
	if rows.Next() {
		res := rows.Scan(&check)
		if res != nil {
			return -1, fmt.Errorf("ID検索エラー: %v", err)
		}
	}
	func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	if check == true {
		res, err := searchCategoryId(db, category)
		if err != nil {
			return -1, fmt.Errorf("カテゴリーが見つかりませんでした: %v", err)
		}
		return res, err
	}
	createdId, err := addCategoryId(db, category)
	if err != nil {
		return -1, fmt.Errorf("カテゴリーを追加できませんでした: %v", err)
	}
	return createdId, err
}
