package main

import (
	"fmt"
	"strings"
	"time"

	"./parser"
	"./pdtconfig"
	"./request"

	"github.com/kr/pretty"
)

var (
	pdtDebug = false
)

func init() {
	err := pdtconfig.Config.Init()
	if err != nil {
		fmt.Println("config.json error!")
		panic(err)
	}

	// var (
	// 	reqHeader = map[string]string{
	// 		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:63.0) Gecko/20100101 Firefox/63.0",
	// 		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	// 		"Accept-Language":           "en-US,en;q=0.8,zh-CN;q=0.5,zh;q=0.3",
	// 		"Accept-Encoding":           "gzip, deflate",
	// 		"DNT":                       "1",
	// 		"Connection":                "keep-alive",
	// 		"Upgrade-Insecure-Requests": "1",
	// 		"Cache-Control":             "max-age=0",
	// 	}
	// )

	request.Params.Socks5Proxy = pdtconfig.Config.Socks5Proxy()
	request.Params.FakeHeaders = pdtconfig.Config.FakeHeaders()
	request.Params.Debug = pdtconfig.Config.HTTPDebug()
	pdtDebug = pdtconfig.Config.Debug()

	parser.Params.PdSiteHost = pdtconfig.Config.PdSite()
	parser.Params.PdURLPage = pdtconfig.Config.PdURLPage()
	parser.Params.UserLoginStr = pdtconfig.Config.UserLoginStr()
	parser.Params.TorrItemInfo = pdtconfig.Config.TorrItemInfo()

	if pdtDebug {
		fmt.Println(pdtDebug)
		fmt.Println(request.Params.Debug)
		fmt.Println(request.Params.Socks5Proxy)
		fmt.Println(pdtconfig.Config.Debug())
		fmt.Println(request.Params.FakeHeaders)

		fmt.Println(parser.Params.PdSiteHost)
		fmt.Println(parser.Params.PdURLPage)
		fmt.Println(parser.Params.UserLoginStr)
		fmt.Println(parser.Params.TorrItemInfo)
	}
}

func main() {

	// 装入需要下载的Torrent 关键字列表，存入torrDownMap中
	torrDownMap, err := pdtconfig.LoadTorrentInfo(pdtconfig.Config.TorrListFile())
	if err != nil {
		panic(err)
	} else if pdtDebug {
		for torrTitle := range torrDownMap {
			fmt.Println(torrTitle, torrDownMap[torrTitle])
		}
	}

	torrSize := len(torrDownMap)
	loopCount := 1

	for {
		var downedTorrentCount, leftoverTorrent int
		var downedMsg, leftoverMsg string
		for title := range torrDownMap {
			if torrDownMap[title] {
				downedTorrentCount++
				downedMsg += fmt.Sprintf("[%02d]%s\n", downedTorrentCount, title)
			} else {
				leftoverTorrent++
				leftoverMsg += fmt.Sprintf("[%02d]%s\n", leftoverTorrent, title)
			}
		}

		fmt.Printf("[%s] 第 %d 次检查页面\n已下载了[%d]个种子文件\n%s尚有[%d]个种子文件未下载\n%s",
			time.Now().Format("2006-01-02 15:04:05"),
			loopCount,
			downedTorrentCount,
			downedMsg,
			torrSize-downedTorrentCount,
			leftoverMsg,
		)

		loopCount++
		torrentInfoMap, err := parser.Params.GetPDTList(0)
		if err != nil {
			fmt.Printf("获取页面出错: %s\n", err)
		} else {
			for title, url := range torrentInfoMap {
				fmt.Printf(title)
				for torrTitle := range torrDownMap {
					if torrDownMap[torrTitle] == false {
						if strings.Contains(title, torrTitle) {
							fmt.Printf("\n---下载: [%d/%d]---\n", downedTorrentCount+1, torrSize)
							fmt.Println(torrTitle)
							fmt.Println(url)

							torrDownMap[torrTitle] = true

							err := request.DownloadFile(torrTitle+".torrent", url)
							if err == nil {
								downedTorrentCount++
							} else {
								fmt.Println("下载出错...")
							}
						}
					}
				}
			}
			if pdtDebug {
				pretty.Println(torrDownMap)
			}
		}

		if downedTorrentCount >= torrSize {
			fmt.Println("\n完成全部下载...")
			break
		}

		fmt.Printf("\n---等待 %d 分钟后的第 %d 次检查...\n\n", pdtconfig.Config.SleepTime(), loopCount)
		time.Sleep(time.Duration(pdtconfig.Config.SleepTime()*60) * time.Second)
	}
}
