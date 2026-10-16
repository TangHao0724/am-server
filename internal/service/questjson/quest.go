package questjson

import (
	"encoding/json"
	"fmt"
)

type Quests struct {
	AllOptionsStageSortID []uint         `json:"all_options_stage_sort_id"`
	Quests                []Quest        `json:"quests"`
	Variable              []SettingValue `json:"variable"`
}
type Quest struct {
	ID          uint           `json:"id"`
	QuestTitle  string         `json:"quest_title"`
	Description *string        `json:"description,omitempty"`
	Setting     []SettingValue `json:"setting,omitempty"`
	Options     []Option       `json:"options"`
}

type Option struct {
	ID             uint             `json:"id"`
	IsCreatorCheck bool             `json:"is_creator_check"`
	Title          string           `json:"title"`
	Answer         OptionAnswer     `json:"answer"`
	Response       []OptionResponse `json:"response"`
}

func (o *Option) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID             uint            `json:"id"`
		IsCreatorCheck bool            `json:"is_creator_check"`
		Title          string          `json:"title"`
		Answer         json.RawMessage `json:"answer"`
		Response       json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode option: %w", err)
	}
	answerUm, err := UnmarshalOptionAnswer(raw.Answer)
	if err != nil {
		return fmt.Errorf("decode option answer: %w", err)
	}
	responses, err := UnmarshalOptionResponses(raw.Response)
	if err != nil {
		return fmt.Errorf("decode option responses: %w", err)
	}
	o.ID = raw.ID
	o.IsCreatorCheck = raw.IsCreatorCheck
	o.Title = raw.Title
	o.Answer = answerUm
	o.Response = responses

	return nil
}

// 設定"設定" 以及單一數值變數
type SettingValue struct {
	Title string `json:"title"`
	Set   int    `json:"set"`
}

type GameState struct {
	GameVariables map[string]int

	// StageID -> Variables
	StageVariables map[uint]map[string]int

	CurrentStageID uint
	CurrentQuestID uint
}
