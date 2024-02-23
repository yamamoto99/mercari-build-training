package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"imageName"`
}

func copyfile(img *multipart.FileHeader) (string, error) {
	// 既存のimgファイルを開く
	originalFile, err := img.Open()
	if err != nil {
		return "", fmt.Errorf("元画像ファイルの読み込みに失敗しました: %v", err)
	}
	// ファイルをhash化
	hash := sha256.New()
	if _, err := io.Copy(hash, originalFile); err != nil {
		return "", fmt.Errorf("hash値の生成に失敗しました: %v", err)
	}
	_ = originalFile.Close()
	// 拡張子をつけるとともにsが[]byteなのでstringに変換
	newFileName := hex.EncodeToString(hash.Sum(nil)) + filepath.Ext(img.Filename)
	// ファイルパスを指定して名前をhash化したファイルを生成
	newFile, err := os.Create("images/" + newFileName)
	if err != nil {
		return "", fmt.Errorf("新しい画像ファイルの生成に失敗しました: %v", err)
	}
	// 既存のimgファイルを開く
	originalFile, err = img.Open()
	if err != nil {
		return "", fmt.Errorf("元画像ファイルの読み込みに失敗しました: %v", err)
	}
	// 新しいファイルに既存のデータをコピー
	if _, err := io.Copy(newFile, originalFile); err != nil {
		return "", fmt.Errorf("ファイルのコピーに失敗しました: %v", err)
	}
	_ = originalFile.Close()
	return newFileName, err
}

func connectDB(c echo.Context) *sql.DB {
	// dbOpen
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		c.Logger().Fatalf("DB接続エラー: %v", err)
	}
	return db
}

func jsonconv(rows *sql.Rows) ([]byte, error) {
	items := new(Items)
	for rows.Next() {
		item := Item{}
		if err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.ImageName); err != nil {
			return nil, err
		}
		items.Items = append(items.Items, item)
	}
	// json.MarshalでJSON形式に変換
	output, err := json.Marshal(&items)
	if err != nil {
		return nil, err
	}
	//// terminal上で末尾改行してない時に自動改行の%が付与されるのを防ぐため、あらかじめ改行を付与
	res := append(output, byte('\n'))

	return res, err
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
	c.Logger().Debugf("Receive item: %s", name)
	c.Logger().Debugf("Receive category: %s", category)
	c.Logger().Debugf("Receive imageFile: %s", img.Filename)
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
	c.Logger().Infof(name)
	c.Logger().Infof(strconv.FormatInt(categoryId, 10))
	c.Logger().Infof(newFileName)
	// ステートメントを用いて書き込み
	if _, err = stmt.Exec(name, categoryId, newFileName); err != nil {
		c.Logger().Fatalf("SQL書き込みエラー: %v", err)
	}

	// ▽JSONに保存▽

	//item := Item{Name: name, Category: category, ImageName: newFileName}
	//// jsonファイル読み込み
	//inJsonItems, err := os.ReadFile("items.json")
	//if err != nil {
	//	c.Logger().Fatalf("JSONデータを読み込めませんでした: %v", err)
	//}
	//var inItems Items
	//// 読み込んだファイルが空ではない時JSONを構造体に変換
	//if len(inJsonItems) != 0 {
	//	err = json.Unmarshal(inJsonItems, &inItems)
	//	if err != nil {
	//		c.Logger().Fatalf("JSONファイルを構造体に変換できませんでした: %v", err)
	//	}
	//}
	//// 構造体にformの値を追加
	//inItems.Items = append(inItems.Items, item)
	//// json.MarshalでJSON形式に変換
	//output, err := json.Marshal(&inItems)
	//if err != nil {
	//	c.Logger().Fatalf("JSON生成中にエラーが発生しました: %v", err)
	//}
	//// jsonファイルを作成、すでにある場合はクリアして開く
	//file, err := os.Create("items.json")
	//if err != nil {
	//	c.Logger().Fatalf("JSONファイルを開けませんでした: %v", err)
	//}
	//// ファイルclose
	//defer func(file *os.File) {
	//	_ = file.Close()
	//}(file)
	//// jsonファイル書き込み
	//if _, err = file.Write(output); err != nil {
	//	c.Logger().Fatalf("JSONファイル書き込み中にエラーが発生しました: %v", err)
	//}
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
	// ▽JSONから取得▽

	//// jsonファイル読み込み
	//afterItems, err := os.ReadFile("items.json")
	//if err != nil {
	//	c.Logger().Fatalf("JSONデータを読み込めませんでした: %v", err)
	//}

	// 取得したのがbyte形式だったのでJSON->JSONBlobに変更
	return c.JSONBlob(http.StatusOK, res)
}

//func getItemById(c echo.Context) error {
//	// jsonファイル読み込み
//	inJsonItems, err := os.ReadFile("items.json")
//	if err != nil {
//		c.Logger().Fatalf("JSONデータを読み込めませんでした: %v", err)
//	}
//	var inItems Items
//	// 読み込んだファイルが空の時早期return
//	if len(inJsonItems) == 0 {
//		res := Response{Message: "itemはまだ登録されていません"}
//		return c.JSON(http.StatusOK, res)
//	}
//	err = json.Unmarshal(inJsonItems, &inItems)
//	if err != nil {
//		c.Logger().Fatalf("JSONファイルを構造体に変換できませんでした: %v", err)
//	}
//	index, err := strconv.Atoi(c.Param("item_id"))
//	if err != nil {
//		c.Logger().Fatalf("idの取得に失敗しました: %v", err)
//	}
//	if len(inItems.Items) <= index {
//		res := Response{Message: "存在しないIDです"}
//		return c.JSON(http.StatusOK, res)
//	}
//	res := inItems.Items[index]
//	return c.JSON(http.StatusOK, res)
//}

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

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/search", searchItem)
	//e.GET("/items/:item_id", getItemById)
	e.GET("/image/:imageId", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
