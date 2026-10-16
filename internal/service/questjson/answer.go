package questjson

import (
	"strings"
)

type OptionAnswer interface {
	IsAnswer()
}

// 給予字串，若Isfuzzy true 。會開啟模糊許可(使用者輸入字串包含AnsStringList內字串及判定成功)
type StringAnswer struct {
	Type          string   `json:"type"`
	AnsStringList []string `json:"ans_string_list"`
	IsFuzzy       bool     `json:"is_fuzzy"`
}

func (StringAnswer) IsAnswer() {}

func (ans *StringAnswer) Validate(request string) bool {
	request = strings.ToLower(strings.TrimSpace(request))

	for _, answer := range ans.AnsStringList {
		answer = strings.ToLower(strings.TrimSpace(answer))

		if answer == "" {
			continue
		}

		if ans.IsFuzzy {
			if strings.Contains(request, answer) {
				return true
			}
		} else {
			if request == answer {
				return true
			}
		}
	}
	return false
}

// 給予選擇。可以在quest的setting 選擇是否多選
type SelectAnswer struct {
	Type  string `json:"type"`
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (ans *SelectAnswer) Validate() bool {
	return true
}

func (SelectAnswer) IsAnswer() {}

// 給予數值對list，若使用者在數值中正確
type IntAnser struct {
	Type      string   `json:"type"`
	ID        int      `json:"id"`
	RangeList [][2]int `json:"range_list"`
}

func (ans *IntAnser) Validate(request int) bool {
	for _, intPair := range ans.RangeList {
		if request >= intPair[0] && request <= intPair[1] {
			return true
		}
	}
	return false
}
func (IntAnser) IsAnswer() {}

// 使用者確認，若遇到這option IsCheck true。直接回傳RawMessage
type CreatorCheckAnser struct {
	Type    string `json:"type"`
	IsCheck bool   `json:"is_check"`
}

func (ans *CreatorCheckAnser) Validate(request int) bool {
	// 1. 額外發起一個操作，在creator_checks 建立一筆新資料
	// 2. 讓遊戲進度暫停在這裡。(停止user對進度的更新)
	// 3. 當creator更新標單內容，完成操作。系統對使用者發出通知。
	// 4. 當玩家更新關卡頁面、或者定時發起主動抓取creator_checks 資料的API check內容。
	// 如果抓取到，回傳true。
	// 總結。暫時不時做(其他功能優先)
	return false
}
func (CreatorCheckAnser) IsAnswer() {}
