package main

import (
	"encoding/json"
	"fmt"
	"regexp"
)

func ExtractJSON(input string) string {
	// 定义正则表达式模式，用于匹配 JSON 格式的字符串
	pattern := `(\{[^{}]*\})`

	// 编译正则表达式
	regex := regexp.MustCompile(pattern)

	// 查找匹配的子字符串
	matches := regex.FindStringSubmatch(input)

	// 如果找到匹配项，则返回第一个匹配项（JSON 字符串）
	if len(matches) > 1 {
		return matches[1]
	}

	// 如果未找到匹配项，返回空字符串或者其他你认为合适的值
	return ""
}

func main() {
	input := `
以下是从文档中提取的关键信息，并以json格式返回：
json
{
  "关键信息": [
    "污蔑曹云金三件事",
    "相声比赛拿第一名，老郭让退赛",
    "老郭说没收学费，金子拿出交款发票",
    "污蔑曹云金离开三年，没发一个短信",
    "郭德纲说冤枉你的人，比你知道有多冤枉",
    "做师父不厚道，让徒弟背锅",
    "曹云金有证据",
    "曹云金被广大网友冠名为“反职场霸凌第一人”"
  ]
}`
	jsonStr := ExtractJSON(input)
	var res map[string][]string
	err := json.Unmarshal([]byte(jsonStr), &res)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)
	fmt.Println(res["关键信息"])
}
