package proxy

import "regexp"

// RegionPattern 地区匹配模式
type RegionPattern struct {
	Name    string         // 分组名称，如 "🇭🇰 香港节点"
	Icon    string         // 图标
	Pattern *regexp.Regexp // 匹配正则
}

// RegionPatterns 地区正则表达式（参考 sing-box-subscribe）
var RegionPatterns = []RegionPattern{
	{
		Name:    "🇭🇰 香港节点",
		Icon:    "🇭🇰",
		Pattern: regexp.MustCompile(`(?i)香港|沪港|呼港|中港|HKT|HKBN|HGC|WTT|CMI|穗港|广港|京港|🇭🇰|HK|Hongkong|Hong Kong|HongKong|HONG KONG`),
	},
	{
		Name:    "🇨🇳 台湾节点",
		Icon:    "🇨🇳",
		Pattern: regexp.MustCompile(`(?i)台湾|台灣|臺灣|台北|台中|新北|彰化|CHT|HINET|🇨🇳|TW|Taiwan|TAIWAN`),
	},
	{
		Name:    "🇸🇬 新加坡节点",
		Icon:    "🇸🇬",
		Pattern: regexp.MustCompile(`(?i)新加坡|狮城|獅城|沪新|京新|泉新|穗新|深新|杭新|广新|廣新|滬新|🇸🇬|SG|Singapore|SINGAPORE`),
	},
	{
		Name:    "🇯🇵 日本节点",
		Icon:    "🇯🇵",
		Pattern: regexp.MustCompile(`(?i)日本|东京|東京|大阪|埼玉|京日|苏日|沪日|广日|上日|穗日|川日|中日|泉日|杭日|深日|🇯🇵|JP|Japan|JAPAN`),
	},
	{
		Name:    "🇺🇸 美国节点",
		Icon:    "🇺🇸",
		Pattern: regexp.MustCompile(`(?i)美国|美國|京美|硅谷|凤凰城|洛杉矶|西雅图|圣何塞|芝加哥|哥伦布|纽约|广美|🇺🇸|US|USA|America|United States`),
	},
	{
		Name:    "🇰🇷 韩国节点",
		Icon:    "🇰🇷",
		Pattern: regexp.MustCompile(`(?i)韩国|韓國|首尔|首爾|韩|韓|春川|🇰🇷|KOR|KR|Korea`),
	},
	{
		Name:    "🇬🇧 英国节点",
		Icon:    "🇬🇧",
		Pattern: regexp.MustCompile(`(?i)英国|英國|伦敦|🇬🇧|UK|England|United Kingdom|Britain`),
	},
	{
		Name:    "🇩🇪 德国节点",
		Icon:    "🇩🇪",
		Pattern: regexp.MustCompile(`(?i)德国|德國|法兰克福|🇩🇪|DE|GER|German|GERMAN`),
	},
	{
		Name:    "🇫🇷 法国节点",
		Icon:    "🇫🇷",
		Pattern: regexp.MustCompile(`(?i)法国|法國|巴黎|🇫🇷|FR|France`),
	},
	{
		Name:    "🇷🇺 俄罗斯节点",
		Icon:    "🇷🇺",
		Pattern: regexp.MustCompile(`(?i)俄罗斯|俄羅斯|毛子|俄国|🇷🇺|RU|RUS|Russia`),
	},
	{
		Name:    "🇮🇳 印度节点",
		Icon:    "🇮🇳",
		Pattern: regexp.MustCompile(`(?i)印度|孟买|🇮🇳|IN|IND|India|Mumbai`),
	},
	{
		Name:    "🇦🇺 澳大利亚节点",
		Icon:    "🇦🇺",
		Pattern: regexp.MustCompile(`(?i)澳大利亚|澳洲|墨尔本|悉尼|🇦🇺|AU|Australia|Sydney`),
	},
	{
		Name:    "🇨🇦 加拿大节点",
		Icon:    "🇨🇦",
		Pattern: regexp.MustCompile(`(?i)加拿大|蒙特利尔|温哥华|多伦多|楓葉|枫叶|🇨🇦|CA|CAN|Canada|CANADA`),
	},
	{
		Name:    "🇳🇱 荷兰节点",
		Icon:    "🇳🇱",
		Pattern: regexp.MustCompile(`(?i)荷兰|荷蘭|阿姆斯特丹|🇳🇱|NL|Netherlands|Amsterdam`),
	},
	{
		Name:    "🇹🇷 土耳其节点",
		Icon:    "🇹🇷",
		Pattern: regexp.MustCompile(`(?i)土耳其|伊斯坦布尔|🇹🇷|TR|TUR|Turkey`),
	},
	{
		Name:    "🇹🇭 泰国节点",
		Icon:    "🇹🇭",
		Pattern: regexp.MustCompile(`(?i)泰国|泰國|曼谷|🇹🇭|TH|Thailand`),
	},
	{
		Name:    "🇻🇳 越南节点",
		Icon:    "🇻🇳",
		Pattern: regexp.MustCompile(`(?i)越南|胡志明市|🇻🇳|VN|Vietnam`),
	},
	{
		Name:    "🇵🇭 菲律宾节点",
		Icon:    "🇵🇭",
		Pattern: regexp.MustCompile(`(?i)菲律宾|菲律賓|🇵🇭|PH|Philippines`),
	},
	{
		Name:    "🇲🇾 马来西亚节点",
		Icon:    "🇲🇾",
		Pattern: regexp.MustCompile(`(?i)马来西亚|马来|馬來|🇲🇾|MY|Malaysia|MALAYSIA`),
	},
	{
		Name:    "🇮🇩 印尼节点",
		Icon:    "🇮🇩",
		Pattern: regexp.MustCompile(`(?i)印尼|印度尼西亚|雅加达|🇮🇩|ID|IDN|Indonesia`),
	},
	{
		Name:    "🇧🇷 巴西节点",
		Icon:    "🇧🇷",
		Pattern: regexp.MustCompile(`(?i)巴西|圣保罗|🇧🇷|BR|Brazil`),
	},
	{
		Name:    "🇦🇷 阿根廷节点",
		Icon:    "🇦🇷",
		Pattern: regexp.MustCompile(`(?i)阿根廷|🇦🇷|AR|Argentina`),
	},
	{
		Name:    "🇦🇪 阿联酋节点",
		Icon:    "🇦🇪",
		Pattern: regexp.MustCompile(`(?i)阿联酋|迪拜|🇦🇪|AE|Dubai|United Arab Emirates`),
	},
	{
		Name:    "🇿🇦 南非节点",
		Icon:    "🇿🇦",
		Pattern: regexp.MustCompile(`(?i)南非|约翰内斯堡|🇿🇦|ZA|South Africa`),
	},
	{
		Name:    "🇲🇽 墨西哥节点",
		Icon:    "🇲🇽",
		Pattern: regexp.MustCompile(`(?i)墨西哥|🇲🇽|MX|MEX|MEXICO`),
	},
}

// ClassifyNodesByRegion 根据节点名称分类到各地区
// 返回 map[地区名][]节点名
func ClassifyNodesByRegion(nodeNames []string) map[string][]string {
	result := make(map[string][]string)
	classified := make(map[string]bool) // 记录已分类的节点

	for _, region := range RegionPatterns {
		var matched []string
		for _, name := range nodeNames {
			if region.Pattern.MatchString(name) {
				matched = append(matched, name)
				classified[name] = true
			}
		}
		if len(matched) > 0 {
			result[region.Name] = matched
		}
	}

	// 未分类的节点放到"其他节点"
	var others []string
	for _, name := range nodeNames {
		if !classified[name] {
			others = append(others, name)
		}
	}
	if len(others) > 0 {
		result["🌍 其他节点"] = others
	}

	return result
}

// GetRegionNames 获取有节点的地区名称列表（按定义顺序）
func GetRegionNames(nodeNames []string) []string {
	classified := ClassifyNodesByRegion(nodeNames)
	var names []string

	for _, region := range RegionPatterns {
		if _, ok := classified[region.Name]; ok {
			names = append(names, region.Name)
		}
	}

	// 如果有其他节点，添加到最后
	if _, ok := classified["🌍 其他节点"]; ok {
		names = append(names, "🌍 其他节点")
	}

	return names
}
