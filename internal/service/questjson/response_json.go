package questjson

import (
	"encoding/json"
	"fmt"
)

func UnmarshalOptionResponse(data []byte) (OptionResponse, error) {
	var header struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("decode response type: %w", err)
	}

	if header.Type == "" {
		return nil, fmt.Errorf("response type is required")
	}

	var response OptionResponse

	switch header.Type {
	case "set_value":
		decoded := &SetValueResponse{}
		if err := json.Unmarshal(data, decoded); err != nil {
			return nil, fmt.Errorf("decode %s response: %w", header.Type, err)
		}
		response = decoded
	case "jump_quest":
		decoded := &DriectQuestResponse{}
		if err := json.Unmarshal(data, decoded); err != nil {
			return nil, fmt.Errorf("decode %s response: %w", header.Type, err)
		}
		response = decoded
	case "jump_stage":
		decoded := &DriectStagesResponse{}
		if err := json.Unmarshal(data, decoded); err != nil {
			return nil, fmt.Errorf("decode %s response: %w", header.Type, err)
		}
		response = decoded
	case "check_variable":
		var raw struct {
			Type           string          `json:"type"`
			Order          int             `json:"order"`
			TargetVariable string          `json:"target_variable"`
			RangeList      [][2]int        `json:"range_list"`
			Response       json.RawMessage `json:"response"`
		}

		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("decode %s response: %w", header.Type, err)
		}
		responses, err := UnmarshalOptionResponses(raw.Response)
		if err != nil {
			return nil, fmt.Errorf("decode nested responses: %w", err)
		}
		decoded := &CheckVariableResponse{
			Type:           raw.Type,
			Order:          raw.Order,
			TargetVariable: raw.TargetVariable,
			RangeList:      raw.RangeList,
			Response:       responses,
		}
		response = decoded
	default:
		return nil, fmt.Errorf("unknown response type %q", header.Type)
	}

	return response, nil
}

func UnmarshalOptionResponses(data []byte) ([]OptionResponse, error) {
	// 1. 將 JSON 陣列拆成多筆 JSON
	var items []json.RawMessage

	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("decode response list: %w", err)
	}

	// 2. 逐筆解析
	responses := make([]OptionResponse, 0, len(items))

	for i, item := range items {
		response, err := UnmarshalOptionResponse(item)
		if err != nil {
			return nil, fmt.Errorf("response[%d]: %w", i, err)
		}

		// 3. 將解析完成的 Response 加入切片
		responses = append(responses, response)
	}

	return responses, nil
}

func MarshalOptionResponse(r OptionResponse) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}

	switch v := r.(type) {
	case *SetValueResponse:
		return json.Marshal(v)
	case SetValueResponse:
		return json.Marshal(v)
	case *DriectQuestResponse:
		return json.Marshal(v)
	case DriectQuestResponse:
		return json.Marshal(v)
	case *DriectStagesResponse:
		return json.Marshal(v)
	case DriectStagesResponse:
		return json.Marshal(v)
	case *CheckVariableResponse:
		return json.Marshal(v)
	case CheckVariableResponse:
		return json.Marshal(v)
	default:
		return nil, fmt.Errorf("unsupported response type %T", r)
	}
}

// NewSetValueResponse 建立設定變數的 Response
func NewSetValueResponse(
	order int,
	targetVariable string,
	setValue int,
) *SetValueResponse {
	return &SetValueResponse{
		Type:           "set_value",
		Order:          order,
		TargetVariable: targetVariable,
		SetValue:       setValue,
	}
}

// NewDirectQuestResponse 建立跳轉 Quest 的 Response
func NewDirectQuestResponse(order int, questID uint) *DriectQuestResponse {
	return &DriectQuestResponse{
		Type:    "jump_quest",
		Order:   order,
		QuestID: questID,
	}
}

// NewDirectStageResponse 建立跳轉 Stage 的 Response
func NewDirectStageResponse(order int, stageSortID uint) *DriectStagesResponse {
	return &DriectStagesResponse{
		Type:        "jump_stage",
		Order:       order,
		StageSortID: stageSortID,
	}
}

// NewCheckVariableResponse 建立變數範圍判斷的 Response
func NewCheckVariableResponse(
	order int,
	targetVariable string,
	rangeList [][2]int,
	responses []OptionResponse,
) *CheckVariableResponse {
	return &CheckVariableResponse{
		Type:           "check_variable",
		Order:          order,
		TargetVariable: targetVariable,
		RangeList:      rangeList,
		Response:       responses,
	}
}
