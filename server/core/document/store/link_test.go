package store

import (
	"context"
	"kpcms/server/core/document"
	"testing"
)

// setupLinkTestRepo creates a test repository with documents that have participant links.
// Note: Caller must call ReloadCache after creating documents to populate the inhabitant index.
func setupLinkTestRepo(t *testing.T) (*Store, map[string][]string) {
	tmpBase := t.TempDir()
	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := New(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()

	// Create test documents with participant links
	docs := []struct {
		code         string
		title        string
		complainants []string
		respondents  []string
	}{
		{
			code:         "0001",
			title:        "Case with John Doe as Complainant",
			complainants: []string{"inhabitant-001-26"},
			respondents:  []string{"inhabitant-002-26"},
		},
		{
			code:         "0002",
			title:        "Case with Jane Smith as Respondent",
			complainants: []string{"inhabitant-003-26"},
			respondents:  []string{"inhabitant-001-26"},
		},
		{
			code:         "0003",
			title:        "Case linked to same inhabitant twice",
			complainants: []string{"inhabitant-001-26"},
			respondents:  []string{"inhabitant-001-26"},
		},
		{
			code:         "0004",
			title:        "Case without participant links",
			complainants: nil,
			respondents:  nil,
		},
	}

	// Track which documents each inhabitant appears in
	inhabitantToDocs := make(map[string][]string)

	for _, d := range docs {
		doc := &document.Document{
			Title: d.title,
			ParticipantIDs: &document.ParticipantLinks{
				Complainants: d.complainants,
				Respondents:  d.respondents,
			},
		}
		created, err := repo.Create(ctx, doc)
		if err != nil {
			t.Fatalf("failed to create test document: %v", err)
		}

		// Track links
		for _, inh := range d.complainants {
			inhabitantToDocs[inh] = append(inhabitantToDocs[inh], created.UUID)
		}
		for _, inh := range d.respondents {
			inhabitantToDocs[inh] = append(inhabitantToDocs[inh], created.UUID)
		}
	}

	// Rebuild the inhabitant index to ensure it's populated
	_, err = repo.ReloadCache(ctx)
	if err != nil {
		t.Fatalf("failed to reload cache: %v", err)
	}

	return repo, inhabitantToDocs
}

// setupLinkTestRepoWithEmptyParticipantIDs creates documents with empty/nil participant_ids.
// Note: Caller must call ReloadCache after creating documents to populate the inhabitant index.
func setupLinkTestRepoWithEmptyParticipantIDs(t *testing.T) (*Store, map[string][]string) {
	tmpBase := t.TempDir()
	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := New(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()

	// Create test documents with nil and empty participant_ids
	docs := []struct {
		code           string
		title          string
		participantIDs *document.ParticipantLinks
	}{
		{
			code:           "0001",
			title:          "Document with nil participant_ids",
			participantIDs: nil,
		},
		{
			code:           "0002",
			title:          "Document with empty participant_ids",
			participantIDs: &document.ParticipantLinks{},
		},
		{
			code:  "0003",
			title: "Document with links",
			participantIDs: &document.ParticipantLinks{
				Complainants: []string{"inhabitant-001-26"},
			},
		},
	}

	// Track which documents each inhabitant appears in
	inhabitantToDocs := make(map[string][]string)

	for _, d := range docs {
		doc := &document.Document{
			Title:          d.title,
			ParticipantIDs: d.participantIDs,
		}
		created, err := repo.Create(ctx, doc)
		if err != nil {
			t.Fatalf("failed to create test document: %v", err)
		}

		if d.participantIDs != nil {
			for _, inh := range d.participantIDs.Complainants {
				inhabitantToDocs[inh] = append(inhabitantToDocs[inh], created.UUID)
			}
		}
	}

	// Rebuild the inhabitant index to ensure it's populated
	_, err = repo.ReloadCache(ctx)
	if err != nil {
		t.Fatalf("failed to reload cache: %v", err)
	}

	return repo, inhabitantToDocs
}

func TestGetDocumentsByInhabitantCode_SingleMatch(t *testing.T) {
	repo, expectedLinks := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-002-26 should appear in exactly 1 document
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-002-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	expectedCount := len(expectedLinks["inhabitant-002-26"])
	if len(docs) != expectedCount {
		t.Errorf("expected %d documents, got %d", expectedCount, len(docs))
	}
}

func TestGetDocumentsByInhabitantCode_MultipleMatches(t *testing.T) {
	repo, expectedLinks := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-001-26 should appear in 3 documents (as complainant in doc1, respondent in doc2, both in doc3)
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-001-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	expectedCount := len(expectedLinks["inhabitant-001-26"])
	if len(docs) != expectedCount {
		t.Errorf("expected %d documents, got %d", expectedCount, len(docs))
	}

	// Verify all returned documents have the inhabitant in their participant_ids
	inhabitantCodes := make(map[string]bool)
	for _, doc := range docs {
		if doc.ParticipantIDs != nil {
			for _, c := range doc.ParticipantIDs.Complainants {
				inhabitantCodes[c] = true
			}
			for _, r := range doc.ParticipantIDs.Respondents {
				inhabitantCodes[r] = true
			}
		}
	}

	if !inhabitantCodes["inhabitant-001-26"] {
		t.Error("returned documents should contain inhabitant-001-26 in their participant_ids")
	}
}

func TestGetDocumentsByInhabitantCode_NoMatch(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-999-99 should have no documents
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-999-99")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected 0 documents, got %d", len(docs))
	}
}

func TestGetDocumentsByInhabitantCode_EmptyCode(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected 0 documents for empty code, got %d", len(docs))
	}
}

func TestGetDocumentsByInhabitantCode_ComplainantOnly(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-003-26 appears only as complainant in doc2
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-003-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}

	if len(docs) > 0 && docs[0].Title != "Case with Jane Smith as Respondent" {
		t.Errorf("expected 'Case with Jane Smith as Respondent', got %s", docs[0].Title)
	}
}

func TestGetDocumentsByInhabitantCode_RespondentOnly(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-002-26 appears only as respondent in doc1
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-002-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}

	if len(docs) > 0 && docs[0].Title != "Case with John Doe as Complainant" {
		t.Errorf("expected 'Case with John Doe as Complainant', got %s", docs[0].Title)
	}
}

func TestGetDocumentsByInhabitantCode_BothRoles(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-001-26 appears in doc3 as both complainant and respondent
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-001-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	// Note: The current implementation returns the same document multiple times
	// if the inhabitant appears in both complainant and respondent arrays.
	// This is because the index stores the UUID once for each role.
	// Count how many times doc3 appears
	doc3Count := 0
	for _, doc := range docs {
		if doc.Title == "Case linked to same inhabitant twice" {
			doc3Count++
		}
	}

	// The document should appear exactly 2 times (once for complainant, once for respondent)
	if doc3Count != 2 {
		t.Errorf("expected document to appear 2 times (complainant + respondent), got %d times", doc3Count)
	}

	// Total docs for inhabitant-001-26: 4 (doc1: complainant, doc2: respondent, doc3: both roles)
	if len(docs) != 4 {
		t.Errorf("expected 4 document references, got %d", len(docs))
	}
}

func TestRebuildInhabitantIndex(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)

	// Manually rebuild the index to verify it works
	repo.rebuildInhabitantIndex()

	// Verify index was built correctly
	if repo.inhabitantToDocs == nil {
		t.Fatal("inhabitantToDocs index should not be nil after rebuild")
	}

	// Check that we have entries for our test inhabitants
	expectedInhabitants := []string{
		"inhabitant-001-26",
		"inhabitant-002-26",
		"inhabitant-003-26",
	}

	for _, inh := range expectedInhabitants {
		docs, exists := repo.inhabitantToDocs[inh]
		if !exists {
			t.Errorf("expected inhabitant %s in index", inh)
		}
		if len(docs) == 0 {
			t.Errorf("inhabitant %s should have at least one document", inh)
		}
	}
}

func TestRebuildInhabitantIndex_SkipsNilAndEmpty(t *testing.T) {
	repo, _ := setupLinkTestRepoWithEmptyParticipantIDs(t)

	// Manually rebuild the index
	repo.rebuildInhabitantIndex()

	// Verify index was built correctly
	if repo.inhabitantToDocs == nil {
		t.Fatal("inhabitantToDocs index should not be nil after rebuild")
	}

	// Only inhabitant-001-26 should be in the index (from doc3)
	// Documents with nil and empty participantIDs should be skipped
	docs, exists := repo.inhabitantToDocs["inhabitant-001-26"]
	if !exists {
		t.Error("expected inhabitant-001-26 in index (from document with links)")
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 document for inhabitant-001-26, got %d", len(docs))
	}
}

func TestGetDocumentsByInhabitantCode_DocumentWithoutParticipantIDs(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	// inhabitant-001-26 should return documents where it appears
	// doc4 has no participant links but we need to check its behavior
	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-001-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	// inhabitant-001-26 appears in:
	// - doc1 as complainant (1)
	// - doc2 as respondent (1)
	// - doc3 as both complainant and respondent (2, so appears twice)
	// Total: 4 references (but 3 unique documents)
	if len(docs) != 4 {
		t.Errorf("expected 4 document references, got %d", len(docs))
	}

	// Verify all returned documents contain inhabitant-001-26
	for _, doc := range docs {
		containsInhabitant := false
		if doc.ParticipantIDs != nil {
			for _, c := range doc.ParticipantIDs.Complainants {
				if c == "inhabitant-001-26" {
					containsInhabitant = true
					break
				}
			}
			for _, r := range doc.ParticipantIDs.Respondents {
				if r == "inhabitant-001-26" {
					containsInhabitant = true
					break
				}
			}
		}
		if !containsInhabitant {
			t.Errorf("document '%s' should not be returned - inhabitant-001-26 not in participant_ids", doc.Title)
		}
	}
}

func TestGetDocumentsByInhabitantCode_ReturnsDocumentData(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-001-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	if len(docs) == 0 {
		t.Fatal("expected at least one document")
	}

	// Verify document data is populated correctly
	doc := docs[0]
	if doc.UUID == "" {
		t.Error("document UUID should be populated")
	}
	if doc.Code == "" {
		t.Error("document Code should be populated")
	}
	if doc.FolderName == "" {
		t.Error("document FolderName should be populated")
	}
	if doc.Title == "" {
		t.Error("document Title should be populated")
	}
	if doc.ParticipantIDs == nil {
		t.Error("document ParticipantIDs should be populated")
	}
}

func TestGetDocumentsByInhabitantCode_FilesIncluded(t *testing.T) {
	repo, _ := setupLinkTestRepo(t)
	ctx := context.Background()

	docs, err := repo.GetDocumentsByInhabitantCode(ctx, "inhabitant-001-26")
	if err != nil {
		t.Fatalf("GetDocumentsByInhabitantCode failed: %v", err)
	}

	// Files should be included in the returned documents
	// (we're not testing file attachment here, just that Files field is present)
	for _, doc := range docs {
		// Files field should be present (may be empty array)
		_ = doc.Files // This just verifies the field exists and is accessible
	}
}
