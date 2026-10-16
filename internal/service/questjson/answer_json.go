package questjson

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func UnmarshalOptionAnswer(data []byte) (OptionAnswer, error) {
	var header struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("decode answer type: %w", err)
	}

	if header.Type == "" {
		return nil, fmt.Errorf("answer type is required")
	}

	var answer OptionAnswer

	switch header.Type {
	case "string":
		answer = &StringAnswer{}
	case "select":
		answer = &SelectAnswer{}
	case "int":
		answer = &IntAnser{}
	case "creator_check":
		answer = &CreatorCheckAnser{}
	default:
		return nil, fmt.Errorf("unknown answer type %q", header.Type)
	}

	if err := json.Unmarshal(data, answer); err != nil {
		return nil, fmt.Errorf("decode %s answer: %w", header.Type, err)
	}

	return answer, nil
}

func MarshalOptionAnswer(a OptionAnswer) ([]byte, error) {
	if a == nil {
		return nil, fmt.Errorf("answer cannot be nil")
	}

	var value any

	switch v := a.(type) {
	case *StringAnswer:
		value = v
	case StringAnswer:
		value = v
	case *SelectAnswer:
		value = v
	case SelectAnswer:
		value = v
	case *IntAnser:
		value = v
	case IntAnser:
		value = v
	case *CreatorCheckAnser:
		value = v
	case CreatorCheckAnser:
		value = v
	default:
		return nil, fmt.Errorf("unsupported answer type %T", a)
	}

	return json.Marshal(value)
}

// 避免不必要的空白並保留 JSON 驗證能力。
func compactJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NewStringAnswer 建立字串答案
func NewStringAnswer(ansStringList []string, isFuzzy bool) *StringAnswer {
	return &StringAnswer{
		Type:          "string",
		AnsStringList: ansStringList,
		IsFuzzy:       isFuzzy,
	}
}

// NewSelectAnswer 建立選擇答案
func NewSelectAnswer(id int, title string) *SelectAnswer {
	return &SelectAnswer{
		Type:  "select",
		ID:    id,
		Title: title,
	}
}

// NewIntAnswer 建立整數答案
func NewIntAnswer(id int, rangeList [][2]int) *IntAnser {
	return &IntAnser{
		Type:      "int",
		ID:        id,
		RangeList: rangeList,
	}
}

// NewCreatorCheckAnswer 建立創作者確認答案
func NewCreatorCheckAnswer(isCheck bool) *CreatorCheckAnser {
	return &CreatorCheckAnser{
		Type:    "creator_check",
		IsCheck: isCheck,
	}
}
