package pdtconfig

import (
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

const (
	// ConfigFilename 配置文件名
	ConfigFilename = "config.json"
)

var (
	// 配置文件全路径
	configFilePath = ConfigFilename

	// Config 配置信息, 由外部调用
	Config = NewConfig(configFilePath)
)

// PdtConfig 存放所有参数设置
type PdtConfig struct {
	torrListFile string
	pdSite       string
	pdURLPage    string
	userLoginStr string
	torrItemInfo string
	debug        bool
	httpDebug    bool
	socks5Proxy  string
	sleepTime    int
	fakeHeaders  map[string]string

	configFilePath string
	configFile     *os.File
	fileMu         sync.Mutex
}

type pdtConfigJSON struct {
	TorrListFile string            `json:"torrent_list_file"`
	PdSite       string            `json:"pd_site"`
	PdURLPage    string            `json:"pd_url_page"`
	UserLoginStr string            `json:"user_login_str"`
	TorrItemInfo string            `json:"torr_item_info"`
	Debug        bool              `json:"debug"`
	HTTPDebug    bool              `json:"http_debug"`
	Socks5Proxy  string            `json:"socks5_proxy"`
	SleepTime    int               `json:"sleep_time"`
	FakeHeaders  map[string]string `json:"fake_headers"`
}

// NewConfig 返回 PdtConfig 指针对象
func NewConfig(configFilePath string) *PdtConfig {
	c := &PdtConfig{
		configFilePath: configFilePath,
	}
	return c
}

// Init 初始化
func (c *PdtConfig) Init() error {
	return c.init()
}

// Socks5Proxy 返回Socks5 Proxy 设置
func (c *PdtConfig) Socks5Proxy() string {
	return c.socks5Proxy
}

// PdSite 返回PD的网址设置
func (c *PdtConfig) PdSite() string {
	return c.pdSite
}

// PdURLPage 返回页面URL地址
func (c *PdtConfig) PdURLPage() string {
	return c.pdURLPage
}

// UserLoginStr 检查用户是否已经登录用的字符串
func (c *PdtConfig) UserLoginStr() string {
	return c.userLoginStr
}

// TorrItemInfo 返回正则字符串
func (c *PdtConfig) TorrItemInfo() string {
	return c.torrItemInfo
}

// TorrListFile 返回TorrListFile文件名
func (c *PdtConfig) TorrListFile() string {
	return c.torrListFile
}

// FakeHeaders 返回Socks5 Proxy 设置
func (c *PdtConfig) FakeHeaders() map[string]string {
	return c.fakeHeaders
}

// Debug 是否允许调试模式
func (c *PdtConfig) Debug() bool {
	return c.debug
}

// HTTPDebug 是否允许Http调试信息输出
func (c *PdtConfig) HTTPDebug() bool {
	return c.httpDebug
}

// SleepTime 间隔时间
func (c *PdtConfig) SleepTime() int {
	return c.sleepTime
}

// 内部函数

func (c *PdtConfig) init() error {
	if c.configFilePath == "" {
		return ErrConfigFileNotExist
	}

	c.initDefaultConfig()
	err := c.loadConfigFromFile()
	if err != nil {
		return err
	}

	// 载入配置
	// 如果 activeUser 已初始化, 则跳过

	return nil
}

// lazyOpenConfigFile 打开配置文件
func (c *PdtConfig) lazyOpenConfigFile() (err error) {
	if c.configFile != nil {
		return nil
	}

	c.fileMu.Lock()
	os.MkdirAll(filepath.Dir(c.configFilePath), 0700)
	c.configFile, err = os.OpenFile(c.configFilePath, os.O_CREATE|os.O_RDWR, 0600)
	c.fileMu.Unlock()

	if err != nil {
		if os.IsPermission(err) {
			return ErrConfigFileNoPermission
		}
		if os.IsExist(err) {
			return ErrConfigFileNotExist
		}
		return err
	}
	return nil
}

// loadConfigFromFile 载入配置
func (c *PdtConfig) loadConfigFromFile() (err error) {
	err = c.lazyOpenConfigFile()
	if err != nil {
		return err
	}

	// 未初始化
	info, err := c.configFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		//err = c.Save()
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	_, err = c.configFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	d := jsoniter.NewDecoder(c.configFile)
	err = d.Decode((*pdtConfigJSON)(unsafe.Pointer(c)))
	if err != nil {
		return ErrConfigContentsParseError
	}
	return nil
}

func (c *PdtConfig) initDefaultConfig() {
	c.torrListFile = "tlist.txt"
	c.debug = false
	c.httpDebug = false
	c.sleepTime = 1800
	c.fakeHeaders = map[string]string{
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:63.0) Gecko/20100101 Firefox/63.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.8,zh-CN;q=0.5,zh;q=0.3",
		"Accept-Encoding":           "gzip, deflate",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Cache-Control":             "max-age=0",
	}
}
