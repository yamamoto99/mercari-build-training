package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
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
	Name     string `json:"name"`
	Category string `json:"category"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive category: %s", category)

	item := Item{Name: name, Category: category}
	// jsonファイル読み込み
	inJsonItems, err := os.ReadFile("items.json")
	if err != nil {
		log.Fatalf("JSONデータを読み込めませんでした: %v", err)
	}
	var inItems Items
	// 読み込んだファイルが空ではない時構造体にJSONデータを格納
	if len(inJsonItems) != 0 {
		err = json.Unmarshal(inJsonItems, &inItems)
		if err != nil {
			log.Fatalf("JSONファイルを構造体に変換できませんでした: %v", err)
		}
	}
	inItems.Items = append(inItems.Items, item)
	// json.MarshalでJSON形式に変換
	output, err := json.Marshal(&inItems)
	if err != nil {
		log.Fatalf("JSON生成中にエラーが発生しました: %v", err)
	}
	file, err := os.Create("items.json")
	if err != nil {
		log.Fatalf("JSONファイルを開けませんでした: %v", err)
	}
	// jsonファイルclose
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("ファイルCloseに失敗しました: %v", err)
		}
	}(file)
	// jsonファイル書き込み
	_, err = file.Write(output)
	if err != nil {
		log.Fatalf("JSONファイル書き込み中にエラーが発生しました: %v", err)
	}
	// jsonファイル読み込み
	afterItems, err := os.ReadFile("items.json")
	if err != nil {
		log.Fatalf("JSONデータを読み込めませんでした: %v", err)
	}
	// terminal上で末尾改行してない時に自動改行の%が付与されるのを防ぐため、あらかじめ改行を付与
	res := append(afterItems, byte('\n'))

	// message := fmt.Sprintf("item received: %s category: %s", name, category)

	// 取得したのがbyte形式だったのでJSON->JSONBlobに変更
	return c.JSONBlob(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
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
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start("127.0.0.1:9000"))
}
