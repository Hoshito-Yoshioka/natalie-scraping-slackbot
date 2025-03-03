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

	// Slack トークンとチャンネルIDの取得
	slackToken := os.Getenv("SLACK_TOKEN")
	channelID := os.Getenv("CHANNEL_ID")

	fmt.Println(slackToken, channelID)

	if slackToken == "" {
		fmt.Println("Error: SLACK_TOKEN is not set")
		os.Exit(1)
	}
	if channelID == "" {
		fmt.Println("Error: CHANNEL_ID is not set")
		os.Exit(1)
	}

	// ニュースの取得
	newsList, err := fetchNews()
	if err != nil {
		postToSlack(slackToken, channelID, "ニュース取得に失敗しました。")
		return
	}

	// Slackへの投稿
	err = postToSlack(slackToken, channelID, formatNewsForSlack(newsList))
	if err != nil {
		log.Fatalf("Failed to post message to Slack: %v", err)
	}
}

// ニュースを取得する関数
func fetchNews() ([]string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// レスポンスの確認を強化
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch the page: %s", resp.Status)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return nil, fmt.Errorf("invalid content type: %s", resp.Header.Get("Content-Type"))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var newsList []string
	doc.Find(".NA_card").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= NEWS_TO_GET {
			return false // 指定数に達したら終了
		}

		// "NA_card_link NA_card_link-tag" を除外して <a> タグを取得
		linkTag := s.Find("a[href]").Last() // 最後の <a> タグを取得
		link, exists := linkTag.Attr("href")
		if !exists || strings.Contains(link, "NA_card_link-tag") {
			return true // 不要なリンクならスキップ
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

		// Slackフォーマットで追加（連番付き）
		newsList = append(newsList, fmt.Sprintf("%d. <%s|%s>", i+1, link, title))

		return true
	})
	return newsList, err
}

// Slackに投稿するためのフォーマットを整える関数
func formatNewsForSlack(newsList []string) string {
	return fmt.Sprintf(":musical_note: 最新ニュースはこちらです :musical_note:\n\n%s\n以上が本日のニュースです！:loudspeaker:",
		strings.Join(newsList, "\n\n"))
}

// Slackへ投稿する関数
func postToSlack(slackToken, channelID, message string) error {
	api := slack.New(slackToken)
	_, _, err := api.PostMessage(channelID, slack.MsgOptionText(message, false))
	return err
}
