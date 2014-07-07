package main

import (
	"encoding/json"
	"flag"
	"os"
	"log"
	"os/user"
	"io/ioutil"
	"github.com/kyokomi/go-gitlab-client/gogitlab"
	"github.com/codegangsta/cli"
	"time"
	"github.com/wsxiaoys/terminal/color"
	"github.com/kyokomi/emoji"
	"fmt"
	"strings"
	"os/exec"
)

type GitlabAccessConfig struct {
	Host     string `json:"host"`
	ApiPath  string `json:"api_path"`
	Token    string `json:"token"`
	IconPath string `json:"icon_path"`
}

func main() {
	app := cli.NewApp()
	app.Version = "0.0.3"
	app.Name = "go-gitlab-notifer"
	app.Usage = "make an explosive entrance"

	app.Flags = []cli.Flag {
		cli.BoolFlag{"gitlab.skip-cert-check", "If set to true, gitlab client will skip certificate checking for https, possibly exposing your system to MITM attack."},
	}

	gitlab := createGitlab()

	app.Commands = []cli.Command{
		{
			Name:      "issue",
			ShortName: "i",
			Usage:     "my issue list",
			Action: func(_ *cli.Context) {
				printGitlabIssues(gitlab)
			},
		},
		{
			Name:      "activity",
			ShortName: "a",
			Usage:     "my activiy list",
			Action: func(_ *cli.Context) {
				printActivity(gitlab)
			},
		},
		{
			Name:      "project",
			ShortName: "p",
			Usage:     "my project list",
			Action: func(_ *cli.Context) {
				printGitlabProjects(gitlab)
			},
		},
		{
			Name:      "tick",
			ShortName: "t",
			Usage:     "my activity list N seconds tick",
			Flags: []cli.Flag{
				cli.IntFlag{"second", 60, "second N."},
			},
			Action: func(c *cli.Context) {
				tickGitlabActivity(gitlab, c.Int("second"))
			},
		},
		{
			// add v0.0.3
			Name:      "events",
			ShortName: "e",
			Usage:     "chice project events",
			Action: func(_ *cli.Context) {
				getProjectIssues(gitlab, 106)
			},
		},
	}

	app.Run(os.Args)
}

func getProjectIssues(gitlab *gogitlab.Gitlab, projectId int) {

	events := gitlab.ProjectEvents(projectId)
	for _, event := range events {

		var iconName string
		switch (event.TargetType) {
		case "Issue":
			iconName = ":beer:"
		default:
			iconName = ":punch:"
		}

		//fmt.Printf("ProjectID[%d] action[%s] targetId[%d] targetType[%s] targetTitle[%s]\n", event.ProductId, event.ActionName,event.TargetId, event.TargetType, event.TargetTitle)
		if event.TargetId != 0 {
			actionText := color.Sprintf("@y[%s]", event.ActionName)
			repositoriesText := color.Sprintf("@c%s(%d)", event.TargetType, event.TargetId)
			userText := color.Sprintf("@c%s", event.Data.UserName)
			titleText := color.Sprintf("@g%s", event.TargetTitle)
			emoji.Println("@{"+iconName+"}", actionText, repositoriesText, userText, titleText)

		} else if event.TargetId == 0 {

			actionText := color.Sprintf("@y[%s]", event.ActionName)
			repositoriesText := color.Sprintf("@c%s", event.Data.Repository.Name)
			userText := color.Sprintf("@c%s", event.Data.UserName)
			var titleText string
			if event.Data.TotalCommitsCount > 0 {
				commitMessage := event.Data.Commits[0].Message
				commitMessage = strings.Replace(commitMessage, "\n\n", "\t", -1)
				titleText = color.Sprintf("@g%s", commitMessage)
			} else if event.Data.Before == "0000000000000000000000000000000000000000" {
				titleText = color.Sprintf("@g%s %s", emoji.Sprint("@{:fire:}"), "create New branch")
			}
			emoji.Println("@{"+iconName+"}", actionText, repositoriesText, userText, titleText)

//			fmt.Println(" \t user   -> ", event.Data.UserName, event.Data.UserId)
//			fmt.Println(" \t author -> ", event.Data.AuthorId)
//
//			fmt.Println(" \t\t name        -> ", event.Data.Repository.Name)
//			fmt.Println(" \t\t description -> ", event.Data.Repository.Description)
//			fmt.Println(" \t\t gitUrl      -> ", event.Data.Repository.GitUrl)
//			fmt.Println(" \t\t pageUrl     -> ", event.Data.Repository.PageUrl)
//
//			fmt.Println(" \t\t totalCount  -> ", event.Data.TotalCommitsCount)
//
//			if event.Data.TotalCommitsCount > 0 {
//				fmt.Println(" \t\t message     -> ", event.Data.Commits[0].Message)
//				fmt.Println(" \t\t time        -> ", event.Data.Commits[0].Timestamp)
//			}
		}
	}
//
//	for _, event := range events {
//
//	}
}

func tickGitlabActivity(gitlab *gogitlab.Gitlab) {

	ch := time.Tick(60 * time.Second)

	lastedFeed, err := gitlab.Activity()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	currentUser, err := gitlab.CurrentUser()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for i := len(lastedFeed.Entries) - 1; i >= 0; i-- {
		feedCommentText := createFeedCommentText(&currentUser, lastedFeed.Entries[i])
		fmt.Println(feedCommentText)
	}

	for now := range ch {
		emoji.Printf("@{:beer:}[%v] tick \n", now)

		feed, err := gitlab.Activity()
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		for i := len(feed.Entries) - 1; i >= 0; i-- {
			feedComment := feed.Entries[i]

			if feedComment.Updated.After(lastedFeed.Entries[0].Updated) {
				feedCommentText := createFeedShortCommentText(feedComment)
				fmt.Println(feedCommentText)
				Notifier(feedCommentText, feedComment.Id, feedComment.Summary, "")
			}
		}
		lastedFeed = feed
	}
}

// 自分の関わってるプロジェクトを取得して一覧表示する
func printGitlabProjects(gitlab *gogitlab.Gitlab) {

	projects, err := gitlab.Projects()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for _, project := range projects {
		fmt.Printf("[%4d] [%20s] (%s)\n", project.Id, project.Name, project.HttpRepoUrl)
	}
}

// 自分のIssueを取得して一覧表示する
func printGitlabIssues(gitlab *gogitlab.Gitlab) {

	issues, err := gitlab.Issues()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for _, issue := range issues {
		fmt.Printf("[%4d] [%s]\n", issue.Id, issue.Title)
	}
}

// Activity（dashboard.atpm）の表示。XMLパースなので情報欠落が辛い感じ。廃止予定
func printActivity(gitlab *gogitlab.Gitlab) {

	feed, err := gitlab.Activity()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	currentUser, err := gitlab.CurrentUser()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for _, feedCommit := range feed.Entries {
		feedCommentText := createFeedCommentText(&currentUser, feedCommit)
		fmt.Println(feedCommentText)
	}
}

// 絵文字付きターミナル用のコメント作成（Activiyベース）
func createFeedCommentText(currentUser *gogitlab.User, feedCommit *gogitlab.FeedCommit) string {

	var iconName string
	if strings.Contains(feedCommit.Title, "commented") {
		iconName = ":speech_balloon:"
	} else if strings.Contains(feedCommit.Title, "pushed") {
		iconName = ":punch:"
	} else if strings.Contains(feedCommit.Title, "closed") {
		iconName = ":white_check_mark:"
	} else if strings.Contains(feedCommit.Title, "opened") {
		iconName = ":fire:"
	} else if strings.Contains(feedCommit.Title, "accepted") {
		iconName = ":atm:"
	} else {
		iconName = ":beer:"
	}

	if currentUser.Name == feedCommit.Author.Name {
		return emoji.Sprint("@{"+iconName+"}", color.Sprintf("@y[%s] %s", feedCommit.Updated.Format(time.ANSIC), feedCommit.Title))
	} else {
		return emoji.Sprint("@{"+iconName+"}", fmt.Sprintf("[%s] %s", feedCommit.Updated.Format(time.ANSIC), feedCommit.Title))
	}
}

// 絵文字付き通知用の簡易コメント作成（Activiyベース）
func createFeedShortCommentText(feedCommit *gogitlab.FeedCommit) string {

	var iconName string
	if strings.Contains(feedCommit.Title, "commented") {
		iconName = ":speech_balloon:"
	} else if strings.Contains(feedCommit.Title, "pushed") {
		iconName = ":punch:"
	} else if strings.Contains(feedCommit.Title, "closed") {
		iconName = ":white_check_mark:"
	} else if strings.Contains(feedCommit.Title, "opened") {
		iconName = ":fire:"
	} else if strings.Contains(feedCommit.Title, "accepted") {
		iconName = ":atm:"
	} else {
		iconName = ":beer:"
	}

	return emoji.Sprint("@{"+iconName+"}", fmt.Sprintf("%s", feedCommit.Title))
}

// Gitlabクライアントを作成する
func createGitlab() *gogitlab.Gitlab {
	config := readGitlabAccessTokenJson()

	// --gitlab.skip-cert-checkを読み込む
	flag.Parse()
	
	return gogitlab.NewGitlab(config.Host, config.ApiPath, config.Token)
}

// アクセストークンを保存してるローカルファイルを読み込んで返却
func readGitlabAccessTokenJson() GitlabAccessConfig {
	usr, err := user.Current()
	if err != nil {
		log.Fatal( err )
		os.Exit(1)
	}

	file, e := ioutil.ReadFile(usr.HomeDir + "/.ggn/config.json")
	if e != nil {
		log.Fatal("Config file error: %v\n", e)
		os.Exit(1)
	}

	var config GitlabAccessConfig
	json.Unmarshal(file, &config)
	config.IconPath = usr.HomeDir + "/.ggn/logo.png"

	return config
}

// 通知のやつ $ brew install terminal-notiferが必要
func Notifier(title, message, subTitle, openUrl string) {
	config := readGitlabAccessTokenJson()
	cmd := exec.Command("terminal-notifier", "-title", title, "-message", message, "-subtitle", subTitle, "-appIcon", config.IconPath, "-open", openUrl)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
