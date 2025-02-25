package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

// ニューススクレイピング関数
func scrapeNatalieNews() ([]string, error) {
	url := "https://natalie.mu/music/news"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("サイトへのアクセスに失敗: %v", err)
	}
	defer resp.Body.Close()

	// goqueryでHTMLを解析
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML解析に失敗: %v", err)
	}

	var newsList []string
	// 取得したニュースを配列に追加
	doc.Find(".NA_card").Each(func(i int, s *goquery.Selection) {
		// 15件以上でループ終了
		if len(newsList) >= 15 {
			return
		}

		link, exists := s.Find("a").Attr("href")
		if !exists || !strings.Contains(link, "/music/news/") {
			return
		}

		title := s.Find(".NA_card_title").Text()
		fullLink := "https://natalie.mu" + link

		// フォーマットしてリストに追加
		newsList = append(newsList, fmt.Sprintf("<%s|%s>", fullLink, title))
	})

	// 実際に追加されたニュースの件数を確認
	if len(newsList) < 15 {
		log.Printf("ニュースは%d件しか取得できませんでした。", len(newsList))
	}

	return newsList, nil
}

// Slackに送信する関数
func postToSlack(newsList []string, slackToken string) error {
	api := slack.New(slackToken)
	message := "🎵 最新ニュースはこちらです 🎵\n"

	// 取得したニュースが15件未満でも、残りを埋める
	for len(newsList) < 15 {
		// ニュースリストに「記事がありません」というエントリを追加
		newsList = append(newsList, "記事がありません")
	}

	// 1〜15の番号を付けて投稿
	var numberedNewsList []string
	for i, news := range newsList[:15] {
		numberedNewsList = append(numberedNewsList, fmt.Sprintf("%d. %s", i+1, news))
	}

	message += strings.Join(numberedNewsList, "\n\n") + "\n\n以上が本日のニュースです！📢"

	_, _, err := api.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionText(message, false))
	return err
}

func main() {
	// .envファイルを読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".envの読み込みに失敗しました")
	}

	// Slackトークン取得
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		log.Fatal("Slackトークンが設定されていません")
	}

	// ニュース取得
	news, err := scrapeNatalieNews()
	if err != nil {
		log.Fatal("ニュース取得エラー:", err)
	}

	// Slackへ送信
	err = postToSlack(news, slackToken)
	if err != nil {
		log.Fatal("Slack送信エラー:", err)
	}

	fmt.Println("ニュースをSlackに送信しました。")
}
