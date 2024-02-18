package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Items struct {
	Items []Item `json:"Items"`
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"imageName"`
}

func copyfile(img *multipart.FileHeader) (string, error) {
	imgName := img.Filename
	// 先頭から(元の長さ-拡張子の長さ)だけ取得
	imgBaseName := imgName[:len(imgName)-len(filepath.Ext(imgName))]
	// 拡張子を除いたファイル名でhash値を生成
	s := sha256.New()
	_, err := io.WriteString(s, imgBaseName)
	if err != nil {
		return "", fmt.Errorf("sha256に変換できませんでした: %v", err)
	}
	// 拡張子をつけるとともにsが[]byteなのでstringに変換
	newFileName := hex.EncodeToString(s.Sum(nil)) + filepath.Ext(imgName)
	// ファイルパスを指定して名前をhash化したファイルを生成
	newFile, err := os.Create("images/" + newFileName)
	if err != nil {
		return "", fmt.Errorf("新しい画像ファイルの生成に失敗しました: %v", err)
	}
	// 既存のimgファイルを開く
	originalFile, err := img.Open()
	if err != nil {
		return "", fmt.Errorf("元画像ファイルの読み込みに失敗しました: %v", err)
	}
	// ファイルclose
	defer func(originalFile multipart.File) {
		_ = originalFile.Close()
	}(originalFile)
	// 新しいファイルに既存のデータをコピー
	if _, err := io.Copy(newFile, originalFile); err != nil {
		return "", fmt.Errorf("ファイルのコピーに失敗しました: %v", err)
	}
	return newFileName, err
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	img, err := c.FormFile("image")
	if err != nil {
		c.Logger().Fatalf("画像の受け取りに失敗しました: %v", err)
	}
	// 入手したファイルから名前をhash化したファイルを生成
	newFileName, err := copyfile(img)
	if err != nil {
		c.Logger().Fatalf("%v", err)
	}

	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive category: %s", category)
	c.Logger().Infof("Receive imageFile: %s", newFileName)

	item := Item{Name: name, Category: category, ImageName: newFileName}
	// jsonファイル読み込み
	inJsonItems, err := os.ReadFile("items.json")
	if err != nil {
		c.Logger().Fatalf("JSONデータを読み込めませんでした: %v", err)
	}
	var inItems Items
	// 読み込んだファイルが空ではない時JSONを構造体に変換
	if len(inJsonItems) != 0 {
		err = json.Unmarshal(inJsonItems, &inItems)
		if err != nil {
			c.Logger().Fatalf("JSONファイルを構造体に変換できませんでした: %v", err)
		}
	}
	// 構造体にformの値を追加
	inItems.Items = append(inItems.Items, item)
	// json.MarshalでJSON形式に変換
	output, err := json.Marshal(&inItems)
	if err != nil {
		c.Logger().Fatalf("JSON生成中にエラーが発生しました: %v", err)
	}
	// jsonファイルを作成、すでにある場合はクリアして開く
	file, err := os.Create("items.json")
	if err != nil {
		c.Logger().Fatalf("JSONファイルを開けませんでした: %v", err)
	}
	// ファイルclose
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	// jsonファイル書き込み
	if _, err = file.Write(output); err != nil {
		c.Logger().Fatalf("JSONファイル書き込み中にエラーが発生しました: %v", err)
	}
	res := Response{Message: "書き込みが正常に完了しました。"}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	// jsonファイル読み込み
	afterItems, err := os.ReadFile("items.json")
	if err != nil {
		c.Logger().Fatalf("JSONデータを読み込めませんでした: %v", err)
	}
	// terminal上で末尾改行してない時に自動改行の%が付与されるのを防ぐため、あらかじめ改行を付与
	res := append(afterItems, byte('\n'))

	// 取得したのがbyte形式だったのでJSON->JSONBlobに変更
	return c.JSONBlob(http.StatusOK, res)
}

func getItemById(c echo.Context) error {
	// jsonファイル読み込み
	inJsonItems, err := os.ReadFile("items.json")
	if err != nil {
		c.Logger().Fatalf("JSONデータを読み込めませんでした: %v", err)
	}
	var inItems Items
	// 読み込んだファイルが空の時早期return
	if len(inJsonItems) == 0 {
		res := Response{Message: "itemはまだ登録されていません"}
		return c.JSON(http.StatusOK, res)
	}
	err = json.Unmarshal(inJsonItems, &inItems)
	if err != nil {
		c.Logger().Fatalf("JSONファイルを構造体に変換できませんでした: %v", err)
	}
	index, err := strconv.Atoi(c.Param("item_id"))
	if err != nil {
		c.Logger().Fatalf("idの取得に失敗しました: %v", err)
	}
	if len(inItems.Items) <= index {
		res := Response{Message: "存在しないIDです"}
		return c.JSON(http.StatusOK, res)
	}
	res := inItems.Items[index]
	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Infof("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
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
	e.GET("/items/:item_id", getItemById)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start("127.0.0.1:9000"))
}
