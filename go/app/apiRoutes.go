package main

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"path"
)

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	if name == "" {
		c.Logger().Fatalf("nameを入力してください")
	}
	category := c.FormValue("category")
	if category == "" {
		c.Logger().Fatalf("カテゴリーを入力してください")
	}
	img, err := c.FormFile("image")
	if err != nil {
		c.Logger().Fatalf("画像の受け取りに失敗しました: %v", err)
	}
	// 入手したファイルから名前をhash化したファイルを生成
	newFileName, err := copyfile(img)
	if err != nil {
		c.Logger().Fatalf("%v", err)
	}
	db := connectDB(c)
	// close処理
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	categoryId, err := checkCategoryId(db, category)
	if err != nil {
		c.Logger().Fatalf("failed id check: %v", err)
	}
	// ステートメントを生成
	stmt, err := db.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		c.Logger().Fatalf("SQL書き込みエラー: %v", err)
	}
	// stmtClose
	defer func(data *sql.Stmt) {
		_ = data.Close()
	}(stmt)
	// ステートメントを用いて書き込み
	if _, err = stmt.Exec(name, categoryId, newFileName); err != nil {
		c.Logger().Fatalf("SQL書き込みエラー: %v", err)
	}
	res := Response{Message: "書き込みが正常に完了しました。"}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	db := connectDB(c)
	// close処理
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	rows, err := db.Query("SELECT items.id, items.name, categories.category_name, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.category_id")
	if err != nil {
		c.Logger().Fatalf("DBから値を取得できませんでした: %v", err)
	}
	res, err := jsonconv(rows)
	// 取得したのがbyte形式だったのでJSONBlob
	return c.JSONBlob(http.StatusOK, res)
}
func searchItem(c echo.Context) error {
	keyword := c.QueryParam("keyword")
	db := connectDB(c)
	// close処理
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	// ステートメントを生成
	stmt, err := db.Prepare("SELECT items.id, items.name, categories.category_name, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.category_id WHERE name LIKE '%' || ? || '%'")
	if err != nil {
		c.Logger().Fatalf("SQL書き込みエラー: %v", err)
	}
	// stmtClose
	defer func(data *sql.Stmt) {
		_ = data.Close()
	}(stmt)
	// DBから取得
	rows, err := stmt.Query(keyword)
	if err != nil {
		c.Logger().Fatalf("DBから値を取得できませんでした: %v", err)
	}
	res, err := jsonconv(rows)
	if err != nil {
		c.Logger().Fatalf("JSONに変換できませんでした: %v", err)
	}
	return c.JSONBlob(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	id := c.Param("imageId")
	db := connectDB(c)
	// close処理
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	stmt, err := db.Prepare("SELECT items.image_name FROM items WHERE id = ?")
	if err != nil {
		c.Logger().Fatalf("stmtを生成できませんでした %v", err)
	}
	c.Logger().Infof(id)
	// Create image path
	var imgPath string
	err = stmt.QueryRow(id).Scan(&imgPath)
	if err != nil {
		c.Logger().Fatalf("imageName取得エラー %v", err)
	}
	imgPath = path.Join(ImgDir, imgPath)
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Infof("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
		c.Logger().Infof(imgPath)
	}

	return c.File(imgPath)
}

func getItemById(c echo.Context) error {
	id := c.Param("item_Id")
	db := connectDB(c)
	// close処理
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	stmt, err := db.Prepare("SELECT * FROM items WHERE id = ?")
	if err != nil {
		c.Logger().Fatalf("stmtを生成できませんでした %v", err)
	}
	defer func(data *sql.Stmt) {
		_ = data.Close()
	}(stmt)
	// DBから取得
	rows, err := stmt.Query(id)
	if err != nil {
		c.Logger().Fatalf("DBから値を取得できませんでした: %v", err)
	}
	res, err := jsonconv(rows)
	if err != nil {
		c.Logger().Fatalf("JSONに変換できませんでした: %v", err)
	}
	return c.JSONBlob(http.StatusOK, res)
}

