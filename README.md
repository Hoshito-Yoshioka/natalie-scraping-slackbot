# natalie-scraping-slackbot

`natalie-scraping-slackbot` は、音楽ナタリーの最新ニュースをスクレイピングし、その内容をSlackチャンネルに自動で投稿するGo言語製のCLIツールです。

## 機能

- 音楽ナタリーの「最新ニュース」ページ (`https://natalie.mu/music/news`) からニュース記事をスクレイピングします。
- 取得したニュースを整形し、Slackに投稿する準備をします。
- 指定されたSlackチャンネルにニュースメッセージを投稿します。
- 投稿するニュースの最大件数を設定できます。

## 必要なもの

- Go 1.23.0 以降
- Slack Bot Token (xoxb-...)
- Slack Channel ID

## 環境変数の設定

以下の環境変数を設定する必要があります。
envファイル、またはgithub actions実行の場合はGithub SECRETSに変数を登録してください。

- `SLACK_TOKEN`: Slack Botのトークン。
- `CHANNEL_ID`: ニュースを投稿するSlackチャンネルのID。

例:

```bash
export SLACK_TOKEN="xoxb-YOUR_SLACK_BOT_TOKEN"
export CHANNEL_ID="YOUR_SLACK_CHANNEL_ID"
```

## インストールと実行方法

1.  このリポジトリをクローンします。

    ```bash
    git clone https://github.com/Hoshito-Yoshioka/natalie-scraping-slackbot.git
    cd natalie-scraping-slackbot
    ```

2.  依存関係を解決します。

    ```bash
    go mod tidy
    ```

3.  環境変数を設定します（上記の「環境変数の設定」を参照）。

4.  プログラムを実行します。
    ```bash
    go run main.go
    ```

## 依存関係

このプロジェクトは以下のGoライブラリに依存しています。

- `github.com/PuerkitoBio/goquery`: HTMLドキュメントの解析とスクレイピングのため。
- `github.com/slack-go/slack`: Slack APIとの連携のため。

## 開発者

[Hoshito-Yoshioka](https://github.com/Hoshito-Yoshioka)

## ライセンス

[LICENSE](LICENSE) (もし存在する場合)
