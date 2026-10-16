package questjson

import "fmt"

func ValidateQuests(data Quests) error {
	questIDs := make(map[uint]struct{})
	allowedStageSortIDs := make(map[uint]struct{})

	// 1. 收集合法 Quest ID
	for _, quest := range data.Quests {
		if err := ValidateQuestAnswerTypes(quest); err != nil {
			return fmt.Errorf("err options size")
		}

		if quest.ID == 0 {
			return fmt.Errorf("quest id cannot be zero")
		}

		if _, exists := questIDs[quest.ID]; exists {
			return fmt.Errorf(
				"duplicate quest id: %d",
				quest.ID,
			)
		}

		questIDs[quest.ID] = struct{}{}
	}

	// 2. 收集允許跳轉的 Stage SortID
	for _, sortID := range data.AllOptionsStageSortID {
		if sortID == 0 {
			return fmt.Errorf("stage sort_id cannot be zero")
		}

		if _, exists := allowedStageSortIDs[sortID]; exists {
			return fmt.Errorf(
				"duplicate stage sort_id: %d",
				sortID,
			)
		}

		allowedStageSortIDs[sortID] = struct{}{}
	}

	// 3. 驗證所有 Quest 的 Response
	for _, quest := range data.Quests {
		for _, option := range quest.Options {
			if err := validateResponses(
				option.Response,
				questIDs,
				allowedStageSortIDs,
			); err != nil {
				return fmt.Errorf(
					"quest[%d] option[%d]: %w",
					quest.ID,
					option.ID,
					err,
				)
			}
		}
	}

	return nil
}
func validateResponses(
	responses []OptionResponse,
	questIDs map[uint]struct{},
	allowedStageSortIDs map[uint]struct{},
) error {

	for index, response := range responses {
		switch r := response.(type) {

		case *DriectQuestResponse:
			if _, exists := questIDs[r.QuestID]; !exists {
				return fmt.Errorf(
					"response[%d]: quest_id %d does not exist",
					index,
					r.QuestID,
				)
			}

		case *DriectStagesResponse:
			if _, exists := allowedStageSortIDs[r.StageSortID]; !exists {
				return fmt.Errorf(
					"response[%d]: stage_sort_id %d is not allowed",
					index,
					r.StageSortID,
				)
			}

		case *CheckVariableResponse:
			// 遞迴檢查巢狀 Response
			if err := validateResponses(
				r.Response,
				questIDs,
				allowedStageSortIDs,
			); err != nil {
				return fmt.Errorf(
					"response[%d] check_variable: %w",
					index,
					err,
				)
			}

		case *SetValueResponse:
			// 此函式暫不驗證變數設定

		default:
			return fmt.Errorf(
				"response[%d]: unsupported response type %T",
				index,
				response,
			)
		}
	}

	return nil
}
func ValidateQuestAnswerTypes(quest Quest) error {
	if len(quest.Options) == 0 {
		return fmt.Errorf("quest %d has no options", quest.ID)
	}

	var answerType string

	for i, option := range quest.Options {
		var currentType string

		switch option.Answer.(type) {
		case *StringAnswer:
			currentType = "string"
		case *IntAnser:
			currentType = "int"
		case *SelectAnswer:
			currentType = "select"
		case *CreatorCheckAnser:
			currentType = "creator_check"
		default:
			return fmt.Errorf(
				"quest %d has unsupported answer type",
				quest.ID,
			)
		}

		if i == 0 {
			answerType = currentType
		} else if currentType != answerType {
			return fmt.Errorf(
				"quest %d has mixed answer types",
				quest.ID,
			)
		}
	}

	if answerType != "select" && len(quest.Options) != 1 {
		return fmt.Errorf(
			"quest %d only select supports multiple options",
			quest.ID,
		)
	}

	return nil
}
