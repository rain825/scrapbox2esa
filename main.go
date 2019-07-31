package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ScrapboxData struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Exported    int    `json:"exported"`
	Pages       []struct {
		Title   string   `json:"title"`
		Created int      `json:"created"`
		Updated int      `json:"updated"`
		Lines   []string `json:"lines"`
	} `json:"pages"`
}

type EsaPostData struct {
	Post Post `json:"post"`
}
type Post struct {
	Name     string   `json:"name"`
	BodyMd   string   `json:"body_md"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Wip      bool     `json:"wip"`
	Message  string   `json:"message"`
	User     string   `json:"user"`
}

var (
	headingRegExp *regexp.Regexp
	listRegExp    *regexp.Regexp
	linkRegExp1   *regexp.Regexp
	linkRegExp2   *regexp.Regexp
	strongRegExp  *regexp.Regexp
	inSiteLink    *regexp.Regexp
	imageRegExp   *regexp.Regexp
	strikeRegExp  *regexp.Regexp
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Error:[usage]TeamName JSONFilePath")
	}

	accessToken := os.Getenv("ESA_ACCESS_TOKEN")
	teamName := os.Args[1]
	dataPath := os.Args[2]
	url := "https://api.esa.io/v1/teams/" + teamName + "/posts"

	client := http.Client{}

	// 正規表現コンパイル
	headingRegExp = regexp.MustCompile(`\[\*\*+ (.+)]`)
	listRegExp = regexp.MustCompile(`^(\t*)\t(.*)$`)
	linkRegExp1 = regexp.MustCompile(`\[(.+) (http[^ ]+)]`)
	linkRegExp2 = regexp.MustCompile(`\[(http[^ ]+) (.*)]`)
	strongRegExp = regexp.MustCompile(`\[\* ([^\]]+)]`)
	inSiteLink = regexp.MustCompile(`#(.+)`)
	imageRegExp = regexp.MustCompile(`\[((https://gyazo\.com.+)|(http.*[png|jpeg|jpg]))]`)
	strikeRegExp = regexp.MustCompile(`\[- ([^\]]+)]`)

	// Jsonファイル読み込み
	rawData, err := ioutil.ReadFile(dataPath)
	if err != nil {
		log.Fatalln(err)
	}

	var data ScrapboxData

	// Jsonデコード
	json.Unmarshal(rawData, &data)

	for _, p := range data.Pages {
		// 初期化
		d := EsaPostData{Post{Name: p.Title, Wip: false, User: "esa_bot", Category: "scrapbox"}}

		fmt.Println(p.Title)
		for i, v := range p.Lines {
			// 先頭のスペースをタブ文字に置換
			spaceNum := 0
			for i := 0; i < len(v); i++ {
				c := v[i]
				if c == ' ' {
					spaceNum++
				} else if '0' <= c && c <= '9' {
					spaceNum = 0
					break
				} else {
					break
				}
			}
			v = strings.TrimLeft(v, " ")
			v = strings.Repeat("\t", spaceNum) + v

			// Scrapbox -> Markdown
			v = inSiteLink.ReplaceAllString(v, "[***$1***](/#)") //esa.ioは作成順がurlに振られるため手動修正
			v = strongRegExp.ReplaceAllString(v, "**$1**")
			v = headingRegExp.ReplaceAllString(v, "## $1")
			v = listRegExp.ReplaceAllString(v, "$1* $2")
			v = linkRegExp1.ReplaceAllString(v, "[$1]($2)")
			v = linkRegExp2.ReplaceAllString(v, "[$2]($1)")
			v = imageRegExp.ReplaceAllString(v, "![image]($1)")
			v = strikeRegExp.ReplaceAllString(v, "~~$1~~")

			// ページの最初はページタイトルと仮定
			if i == 0 {
				d.Post.BodyMd += "# " + v + "\n"
			} else {
				d.Post.BodyMd += v + "\n"
			}
		}

		// Jsonに変換
		body, err := json.Marshal(d)
		if err != nil {
			log.Println(err)
			continue
		}
		// リクエスト生成
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Println(err)
			continue
		}

		// Header付与
		req.Header.Add("Authorization", "Bearer "+accessToken)
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		// fmt.Println(req)
		fmt.Println(res)
	}

}
