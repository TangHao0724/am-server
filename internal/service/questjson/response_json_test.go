package questjson

import (
	"testing"
)

func TestUnmarshalNestedCheckVariable(t *testing.T) {
	data := []byte(`{
        "type": "check_variable",
        "order": 1,
        "target_variable": "score",
        "range_list": [[10, 20]],
        "response": [
            {
                "type": "set_value",
                "order": 2,
                "target_variable": "score",
                "set_value": 100
            },
            {
                "type": "jump_quest",
                "order": 3,
                "quest_id": 5
            }
        ]
    }`)

	got, err := UnmarshalOptionResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	check, ok := got.(*CheckVariableResponse)
	if !ok {
		t.Fatalf("expected *CheckVariableResponse, got %T", got)
	}

	if len(check.Response) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(check.Response))
	}

	if _, ok := check.Response[0].(*SetValueResponse); !ok {
		t.Errorf("response[0]: expected *SetValueResponse, got %T", check.Response[0])
	}

	if _, ok := check.Response[1].(*DriectQuestResponse); !ok {
		t.Errorf("response[1]: expected *DriectQuestResponse, got %T", check.Response[1])
	}
}
