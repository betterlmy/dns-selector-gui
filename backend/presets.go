package backend

import (
	"fmt"

	"github.com/betterlmy/dns-selector/selector"
)

// Preset 表示一个命名的预设方案，包含 DNS 服务器列表和测试域名列表。
type Preset struct {
	Name    string
	Servers []selector.DNSServer
	Domains []string
}

// CNPreset 中国大陆预设方案：
// 包含国内常用 DNS 服务器（UDP + DoT + DoH）和热门域名（仅国内域名）。
var CNPreset = Preset{
	Name:    "cn",
	Servers: selector.GetDefaultCNServers(),
	Domains: cnDomains,
}

// cnDomains 是 CN 预设的测试域名列表，仅包含国内常用域名，去除国外域名。
var cnDomains = []string{
	"douyin.com",
	"kuaishou.com",
	"baidu.com",
	"taobao.com",
	"mi.com",
	"aliyun.com",
	"bilibili.com",
	"jd.com",
	"qq.com",
	"ithome.com",
	"hupu.com",
	"feishu.cn",
	"sohu.com",
	"163.com",
	"sina.com",
	"weibo.com",
	"xiaohongshu.com",
	"douban.com",
	"zhihu.com",
	"youku.com",
	"youdao.com",
	"mp.weixin.qq.com",
	"iqiyi.com",
	"v.qq.com",
	"y.qq.com",
	"www.ctrip.com",
	"autohome.com.cn",
}

// GlobalPreset 全球预设方案：
// 包含国际主流 DNS 服务器（UDP + DoT + DoH）和热门域名。
var GlobalPreset = Preset{
	Name:    "global",
	Servers: selector.GetDefaultGlobalServers(),
	Domains: selector.GetDefaultGlobalDomains(),
}

// GetPreset 根据名称返回对应的预设方案（"cn" 或 "global"）。
// 名称无效时返回错误。
func GetPreset(name string) (*Preset, error) {
	switch name {
	case "cn":
		return &CNPreset, nil
	case "global":
		return &GlobalPreset, nil
	default:
		return nil, fmt.Errorf("未知的预设方案: %q，有效值为 \"cn\" 和 \"global\"", name)
	}
}
