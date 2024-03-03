package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

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
