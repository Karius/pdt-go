package pdtconfig

import (
	"bufio"
	"os"
)

// LoadTorrentInfo 从文件中读取需要下载的torrent标题
func LoadTorrentInfo(filename string) (map[string]bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tListMap := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		tListMap[scanner.Text()] = false
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tListMap, nil
}
