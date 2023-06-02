package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

const (
	DateFormat = "2006-01-02" // 日付のフォーマット
)

// APIからのレスポンスデータの構造
type response struct {
	Date        string `json:"date"`
	Explanation string `json:"explanation"`
	URL         string `json:"url"`
	Title       string `json:"title"`
}

// テンプレートに渡されるビューデータの構造
type viewData struct {
	Date        string	   
	Result      *response  // APIからのレスポンスデータ
}

// APIクライアントの情報を保持する
type apiClient struct {
	apiKey string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	apiKey := os.Getenv("NASA_API_KEY")

	e := echo.New()

	// views ディレクトリ内の静的ファイルをルートパスで公開するように指定する
	e.Static("/", "views")

	t := template.Must(template.ParseFiles("views/index.html"))

	client := apiClient{apiKey: apiKey}

	e.GET("/", client.handleRoot(t))
	e.POST("/", client.handlePost(t))

	fmt.Println("サーバーを起動しました。http://localhost:8080 をブラウザで開いてください。")
	if err := e.Start(":8080"); err != nil {
		log.Fatal("Server error:", err)
	}
}

func (c *apiClient) handleRoot(t *template.Template) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		currentDate := time.Now().Format(DateFormat)
		result, err := c.fetchData(currentDate)
		if err != nil {
			return handleError(ctx, "APIリクエストエラー")
		}

		viewData := viewData{
			Date:        currentDate,
			Result:      result,
		}
		return t.Execute(ctx.Response().Writer, viewData)
	}
}

func (c *apiClient) handlePost(t *template.Template) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		date := ctx.FormValue("date")
		// 指定された日付の情報を取得する
		result, err := c.fetchData(date) 
		if err != nil {
			return handleError(ctx, "APIリクエストエラー")
		}

		viewData := viewData{
			Date:        date,
			Result:      result,
		}
		return t.Execute(ctx.Response().Writer, viewData)
	}
}

// 指定された日付のデータをNASAのAPIから取得する
func (c *apiClient) fetchData(date string) (*response, error) {
	url := fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=%s&date=%s", c.apiKey, date)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := readAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func handleError(ctx echo.Context, message string) error {
	log.Println("エラー:", message)
	return ctx.String(http.StatusInternalServerError, message)
}

func readAll(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
