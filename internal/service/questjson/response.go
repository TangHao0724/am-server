// 回答之後會有的動作
package questjson

type OptionResponse interface {
	IsResponse()
}

// 設定quests中的Value
type SetValueResponse struct {
	Type           string `json:"type"`
	Order          int    `json:"order"`
	TargetVariable string `json:"target_variable"`
	SetValue       int    `json:"set_value"`
}

// func (res *SetValueResponse) action(state *GameState) error {
// 	if _, exists := state.Variables[res.TargetVariable]; !exists {
// 		return fmt.Errorf(
// 			"variable %q not found",
// 			res.TargetVariable,
// 		)
// 	}

// 	state.Variables[res.TargetVariable] = res.SetValue
// 	return nil
// }
func (SetValueResponse) IsResponse() {}

// 前往指定問題
type DriectQuestResponse struct {
	Type    string `json:"type"`
	Order   int    `json:"order"`
	QuestID uint   `json:"quest_id"`
}

func (DriectQuestResponse) IsResponse() {}

// 前往指定關卡
type DriectStagesResponse struct {
	Type        string `json:"type"`
	Order       int    `json:"order"`
	StageSortID uint   `json:"stage_id"`
}

func (DriectStagesResponse) IsResponse() {}

// 設置數值對list，若數值在其中會對target變數進行判斷運算
type CheckVariableResponse struct {
	Type           string           `json:"type"`
	Order          int              `json:"order"`
	TargetVariable string           `json:"target_variable"`
	RangeList      [][2]int         `json:"range_list"`
	Response       []OptionResponse `json:"response"`
}

func (CheckVariableResponse) IsResponse() {}
