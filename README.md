# scrapbox2esa

scrapboxから取得したページをesa.ioにインポートするツールです。

## 使い方
1. Scrapbox Settings>Page Data>Export Pages
2. esa.ioのアクセストークンを取得
3. アクセストークンを環境変数`ESA_ACCESS_TOKEN`に設定
4. `go main.go TeamName PageDataFilePath`


## 参考
markdownへの変換はこちらを参考にしました。[md2sb-online](https://github.com/hashrock/md2sb-online)