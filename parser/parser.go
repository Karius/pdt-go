package parser

import (
	"errors"
	"fmt"
	"regexp"

	"../request"
)

// Params 存方参数设置
type params struct {
	PdSiteHost   string
	PdURLPage    string
	UserLoginStr string
	TorrItemInfo string
}

var (
	// Params 参数设置
	Params = &params{}
)

// // NewParams 创建
// func NewParams() *params {
// 	c := &params{}
// 	return c
// }

// GetPDTList 抓取PD的torrent列表源码
func (c *params) GetPDTList(pagenumber uint) (map[string]string, error) {

	//fmturl := PdSiteHost + PdURLPage
	//url := fmt.Sprintf(fmturl, pagenumber)
	url := fmt.Sprintf(c.PdSiteHost+c.PdURLPage, pagenumber)

	//fmt.Println(url)

	htmlSrc, err := request.GetPageHTML(url)

	if err != nil {
		return nil, err
	}

	isLogin, err := regexp.MatchString(c.UserLoginStr, htmlSrc)
	if err != nil {
		return nil, err
	}
	if isLogin {
		return nil, errors.New("user not login")
	}

	cmap := c.getNewTorrentList(htmlSrc)

	return cmap, nil

}

// --------------------------------------------------------------------------------------------
// 以下为内部函数

// GetNewTorrentList 解析页面内所有torrent的标题和链接并返回
func (c *params) getNewTorrentList(htmlSrc string) map[string]string {
	var cmap map[string]string

	//re2 := regexp.MustCompile(`<td align="left" .*?><a href="javascript:popdetails.*?>(.*?)</a>(.|\n)*?<td .*?</td>(.|\n)*?<td .*?<a href="(.*?)"`)

	re2 := regexp.MustCompile(c.TorrItemInfo)

	matches := re2.FindAllStringSubmatch(htmlSrc, -1)
	// fmt.Println (submatchall)

	//captures:=make([](map[string]string),0)
	names := re2.SubexpNames()

	// fmt.Println (names)

	cmap = make(map[string]string)

	for _, match := range matches {

		title := ""
		downurl := ""

		for pos, val := range match {
			name := names[pos]
			if name == "" {
				continue
			}

			//fmt.Println("+++++++++")
			//fmt.Println(name)
			//fmt.Println(val)
			if name == "title" {
				title = val
			} else if name == "downurl" {
				downurl = fmt.Sprintf("%s/%s", c.PdSiteHost, val)
			}
			// cmap[name]=val
		}
		cmap[title] = downurl
		// captures=append(captures,cmap)
	}

	return cmap
}
