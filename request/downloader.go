package request

import (
	"bytes"
	"io"
	"os"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := Get(url, "", nil)
	if err != nil {
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, bytes.NewReader([]byte(resp)))
	if err != nil {
		return err
	}

	return nil
}

// GetPageHTML 抓取指定页面源码
func GetPageHTML(urladdr string) (string, error) {
	return Get(urladdr, "", nil)
}
