package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/slack-go/slack"
)

const (
	URL         string = "https://natalie.mu/music/news" // 音楽ナタリー"最新ニュース"ページ
	NEWS_TO_GET int    = 15                              // ニュースの取得上限件数
)

func main() {

	// SlackトークンとチャンネルIDの取得
	slackToken := os.Getenv("SLACK_TOKEN")
	channelID := os.Getenv("CHANNEL_ID")

	if slackToken == "" {
		log.Fatal("Error: SLACK_TOKEN is not set")
	}
	if channelID == "" {
		log.Fatal("Error: CHANNEL_ID is not set")
	}

	// ニュースの取得
	newsList, err := fetchNews()
	if err != nil {
		log.Printf("Error: Failed to fetch news: %v", err)
		postNews(slackToken, channelID, fmt.Sprintf("本日のニュース取得に失敗しました... :cry:\n\nエラー: ```%v```", err))
		return
	}

	// ニュースの投稿
	err = postNews(slackToken, channelID, formatNews(newsList))
	if err != nil {
		log.Printf("Error: Failed to post news: %v", err)
		return
	}
}

// ニュースを取得する
func fetchNews() ([]string, error) {
	res, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("ニュースページへのリクエストに失敗しました: %v", err)
	}
	defer res.Body.Close()

	// レスポンスの確認
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ページアクセスに失敗しました。 ステータスコード: %s", res.Status)
	}

	// ドキュメント取得
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var newsList []string
	selection := doc.Find(".NA_card")
	for i := 0; i < selection.Length() && i < NEWS_TO_GET; i++ {
		s := selection.Eq(i)

		// "NA_card_link NA_card_link-tag" を除外して<a>タグを取得
		linkTag := s.Find("a[href]").Last()
		link, exists := linkTag.Attr("href")
		if !exists || strings.Contains(link, "NA_card_link-tag") {
			continue // 不要なリンクをスキップ
		}

		// 記事タイトルを取得
		title := s.Find(".NA_card_title").Text()
		if title == "" {
			title = strings.TrimSpace(s.Find("h3, p").First().Text())
		}

		// リンクを絶対パスに修正
		if !strings.HasPrefix(link, "http") {
			link = "https://natalie.mu" + link
		}

		// ニュースを出力用の配列に追加
		newsList = append(newsList, fmt.Sprintf("%d. <%s|%s>", i+1, link, title))
	}

	// ループ処理でニュースが取得できていなかった場合
	if len(newsList) == 0 {
		return nil, fmt.Errorf("ニュースが見つかりませんでした")
	}

	return newsList, nil
}

// Slackに投稿するためのフォーマットを整える
func formatNews(newsList []string) string {
	return fmt.Sprintf(":musical_note: 最新ニュースはこちらです :musical_note:\n\n%s\n\n以上が本日のニュースです！:loudspeaker:",
		strings.Join(newsList, "\n\n"))
}

// Slackへ投稿
func postNews(slackToken, channelID, message string) error {
	api := slack.New(slackToken)
	_, _, err := api.PostMessage(channelID, slack.MsgOptionText(message, false))
	return err
}
