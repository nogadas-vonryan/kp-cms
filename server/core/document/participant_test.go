package document

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParticipantLinks_JSON(t *testing.T) {
	// Test that ParticipantLinks marshals and unmarshals correctly
	links := &ParticipantLinks{
		Complainants: []string{"inhabitant-001-26", "inhabitant-002-26"},
		Respondents:  []string{"inhabitant-003-26"},
	}

	data, err := json.Marshal(links)
	if err != nil {
		t.Fatalf("Failed to marshal ParticipantLinks: %v", err)
	}

	var unmarshaled ParticipantLinks
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ParticipantLinks: %v", err)
	}

	if len(unmarshaled.Complainants) != 2 {
		t.Errorf("Expected 2 complainants, got %d", len(unmarshaled.Complainants))
	}

	if len(unmarshaled.Respondents) != 1 {
		t.Errorf("Expected 1 respondent, got %d", len(unmarshaled.Respondents))
	}
}

func TestParticipantLinks_Empty(t *testing.T) {
	// Test that empty ParticipantLinks marshals correctly
	links := &ParticipantLinks{}

	data, err := json.Marshal(links)
	if err != nil {
		t.Fatalf("Failed to marshal empty ParticipantLinks: %v", err)
	}

	// Empty object should be valid JSON
	if string(data) != "{}" {
		t.Errorf("Expected empty object, got %s", string(data))
	}
}

func TestParticipantLinks_Nil(t *testing.T) {
	// Test that nil ParticipantIDs is handled correctly in Document
	doc := &Document{
		UUID:       "test-uuid",
		Code:       "001-26",
		FolderName: "case-001-26",
		Title:      "Test Case",
		Fields: map[string]any{
			"complainants": []string{"John Doe"},
			"respondents":  []string{"Jane Smith"},
		},
		ParticipantIDs: nil,
		Files:          nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal document with nil ParticipantIDs: %v", err)
	}

	// Verify that participant_ids is omitted in JSON
	var unmarshaled Document
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal document: %v", err)
	}

	if unmarshaled.ParticipantIDs != nil {
		t.Error("Expected ParticipantIDs to be nil after unmarshal")
	}
}

func TestParticipantLinks_WithData(t *testing.T) {
	// Test that ParticipantIDs with data is included in JSON
	doc := &Document{
		UUID:       "test-uuid",
		Code:       "001-26",
		FolderName: "case-001-26",
		Title:      "Test Case",
		Fields: map[string]any{
			"complainants": []string{"John Doe"},
			"respondents":  []string{"Jane Smith"},
		},
		ParticipantIDs: &ParticipantLinks{
			Complainants: []string{"inhabitant-001-26"},
			Respondents:  []string{"inhabitant-002-26"},
		},
		Files:     nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal document with ParticipantIDs: %v", err)
	}

	// Verify that participant_ids is included in JSON
	var unmarshaled Document
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal document: %v", err)
	}

	if unmarshaled.ParticipantIDs == nil {
		t.Fatal("Expected ParticipantIDs to not be nil after unmarshal")
	}

	if len(unmarshaled.ParticipantIDs.Complainants) != 1 {
		t.Errorf("Expected 1 complainant, got %d", len(unmarshaled.ParticipantIDs.Complainants))
	}

	if len(unmarshaled.ParticipantIDs.Respondents) != 1 {
		t.Errorf("Expected 1 respondent, got %d", len(unmarshaled.ParticipantIDs.Respondents))
	}
}
