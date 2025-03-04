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
	URL         string = "https://natalie.mu/music/news"
	NEWS_TO_GET int    = 15
)

func main() {

	// SlackトークンとチャンネルIDの取得
	slackToken := os.Getenv("SLACK_TOKEN")
	channelID := os.Getenv("CHANNEL_ID")

	if slackToken == "" {
		fmt.Println("Error: SLACK_TOKEN is not set")
		os.Exit(1)
	}
	if channelID == "" {
		fmt.Println("Error: CHANNEL_ID is not set")
		os.Exit(1)
	}

	newsList, err := fetchNews()
	if err != nil {
		postNews(slackToken, channelID, "Error: Failed to fetch news")
		return
	}

	err = postNews(slackToken, channelID, formatNews(newsList))
	if err != nil {
		log.Fatalf("Failed to post message to Slack: %v", err)
	}
}

// ニュースを取得する
func fetchNews() ([]string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// レスポンスの確認
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ページアクセスに失敗しました: %s", resp.Status)
	}

	// ドキュメント取得
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var newsList []string
	doc.Find(".NA_card").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= NEWS_TO_GET {
			return false // 指定数に達したら終了
		}

		// "NA_card_link NA_card_link-tag" を除外して<a>タグを取得
		linkTag := s.Find("a[href]").Last() // 複数の<a>タグが入る場合があるので最後の<a>タグを取得
		link, exists := linkTag.Attr("href")
		if !exists || strings.Contains(link, "NA_card_link-tag") {
			return true // 不要なリンクをスキップ
		}

		// 記事タイトルを取得
		title := s.Find(".NA_card_title").Text()
		if title == "" {
			title = strings.TrimSpace(s.Find("h3, p").First().Text())
		}

		// URLを絶対パスに修正
		if !strings.HasPrefix(link, "http") {
			link = "https://natalie.mu" + link
		}

		// ニュースを出力用の配列に追加（タイトルの前にナンバリングを追加）
		newsList = append(newsList, fmt.Sprintf("%d. <%s|%s>", i+1, link, title))

		return true
	})
	return newsList, err
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
