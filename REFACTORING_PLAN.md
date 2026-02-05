# Plan v2: Folder-Based Database Engine with Cross-Linking

Transform the storage architecture into a unified folder-based database engine with grouped entity folders and explicit cross-linking between documents and inhabitants.

## Vision

**Old Structure (Flat):**
```
/data
├── case-001-25
├── case-999-25
├── case-001-26
```

**New Structure (Grouped):**
```
/data
├── cases/                    # Document folder (cases)
│   ├── case-999-25/
│   │   ├── meta.json
│   │   └── files.json
│   └── case-001-26/
│       ├── meta.json
│       └── files.json
├── inhabitants/              # Document folder (inhabitants)
│   ├── inhabitant-001-26/
│   │   └── inhabitant.json
│   └── inhabitant-002-26/
│       └── inhabitant.json
└── custom/                   # Document folder (future)
    ├── custom-001-24/
    └── custom-002-25/
```

## Priority Order

1. **Main Priority:** Link inhabitants to document cases via explicit IDs
2. **Secondary Priority:** Migrate inhabitants to file-system storage with grouped folder structure

---

## ✅ Phase 0: Folder Restructuring (COMPLETED)

**Goal:** Migrate existing flat `/data/case-XXX-YY` folders to grouped `/data/cases/case-XXX-YY` structure.

### Tasks

1. **✅ Create Migration Script** 
   - Script location: `server/cmd/scripts/restructure/main.go`
   - Move all existing `/data/case-XXX-YY/` folders into `/data/cases/`
   - Create empty `/data/inhabitants/` and `/data/custom/` directories
   - Validate no data loss after migration

2. **Update Document Store Base Path**
   - Modify `server/cmd/server/main.go` to use `{dataPath}/cases` as base path:
     ```go
     // Before
     documentRepository, err := store.New(*dataPath, *backupPath, namingStrategy)
     
     // After
     casesPath := filepath.Join(*dataPath, "cases")
     documentRepository, err := store.New(casesPath, *backupPath, namingStrategy)
     ```

3. **Update Backup Path Logic**
   - Ensure backup operations work with new nested structure
   - Backup should capture entire `/data` folder (all entity types)

4. **Test & Validate**
   - Run existing document tests against new path structure
   - Verify cache reload works correctly
   - Confirm file upload/download paths resolve correctly

### Rollback Strategy
Keep a backup of original `/data` folder before migration. Script includes `--dry-run` mode to preview changes.

---

## Phase 1: Document-Inhabitant Linking (Main Priority)

**Goal:** Enable precise cross-referencing between documents and inhabitants using IDs.

### Tasks

1. **Update Document Schema**
   - Add `participantIds` field to document `meta.json`:
     ```json
     {
       "uuid": "...",
       "code": "001-26",
       "folder_name": "case-001-26",
       "title": "Sample Case",
       "fields": {
         "complainants": ["John Doe"],
         "respondents": ["Jane Smith"],
         "status": "filed"
       },
       "participant_ids": {
         "complainants": ["inhabitant-001-26"],
         "respondents": ["inhabitant-002-26"]
       }
     }
     ```
   - Keep existing `fields.complainants` and `fields.respondents` for display purposes
   - `participant_ids` used exclusively for linking and queries

2. **Add ParticipantIDs to Document Model**
   - Update `server/core/document/model.go`:
     ```go
     type Document struct {
         UUID           string
         Code           string
         FolderName     string
         Title          string
         Fields         map[string]any
         ParticipantIDs *ParticipantLinks  // NEW
         Files          []File
         CreatedAt      time.Time
         UpdatedAt      time.Time
     }
     
     type ParticipantLinks struct {
         Complainants []string `json:"complainants,omitempty"`
         Respondents  []string `json:"respondents,omitempty"`
     }
     ```

3. **Create Name-to-ID Mapping Script**
   - Script location: `server/cmd/scripts/link_participants.go`
   - **Step 1:** Generate mapping report (`--report` mode):
     - Extract all unique participant names from documents
     - Use tokenized matching (reuse logic from `search/aggregator.go`)
     - Output: CSV/JSON showing `name → inhabitant_code` mappings with confidence scores
   - **Step 2:** Apply mappings (`--apply` mode):
     - Read validated mapping file
     - Update each document's `participant_ids` field
     - Log all changes for audit

4. **Matching Algorithm**
   - High confidence (>90%): Auto-match
   - Low confidence: Flag for manual review
   - Use same name variation logic from `generateSearchKeys()` in [aggregator.go](server/core/search/aggregator.go#L118):
     - Full name: "John Dabba Doe"
     - Short name: "John Doe"
     - Formal: "Doe, John"
     - Formal with middle initial: "Doe, John D."

5. **Update Search Aggregator**
   - Modify `SearchByParticipants` to support both:
     - ID-based lookup (when `participant_ids` present)
     - Name-based fallback (for legacy documents without IDs)
   - Add `GetDocumentsByInhabitantCode(code string) ([]*Document, error)` method

6. **Build Reverse Index (Inhabitant → Documents)**
   - Create in-memory index during cache warmup
   - Structure: `map[string][]string` (inhabitantCode → []documentUUID)
   - Derive from document `participant_ids`
   - Rebuild on document cache refresh

### API Considerations
- New endpoint: `GET /api/inhabitants/{code}/documents` - returns all documents linked to an inhabitant
- Existing document endpoints unchanged (backward compatible)

---

## Phase 2: Inhabitant File-System Store (Secondary Priority)

**Goal:** Implement file-based storage for inhabitants in `/data/inhabitants/` folder.

### Tasks

1. **Extend Inhabitant Model**
   - Update `server/core/inhabitant/model.go`:
     ```go
     type Inhabitant struct {
         // New fields for file-based storage
         UUID       string    `json:"uuid"`
         Code       string    `json:"code"`        // e.g., "001-26"
         FolderName string    `json:"folder_name"` // e.g., "inhabitant-001-26"
         CreatedAt  time.Time `json:"created_at"`
         UpdatedAt  time.Time `json:"updated_at"`
         
         // Existing fields (keep all)
         ID                           int64     `json:"-"` // Deprecated, for migration only
         FirstName                    string    `json:"first_name"`
         LastName                     string    `json:"last_name"`
         MiddleName                   string    `json:"middle_name,omitempty"`
         Suffix                       string    `json:"suffix,omitempty"`
         Birthdate                    time.Time `json:"birthdate,omitempty"`
         BirthPlace                   string    `json:"birth_place,omitempty"`
         InhabitantType               string    `json:"inhabitant_type,omitempty"`
         Sex                          string    `json:"sex,omitempty"`
         CivilStatus                  string    `json:"civil_status,omitempty"`
         Citizenship                  string    `json:"citizenship,omitempty"`
         Occupation                   string    `json:"occupation,omitempty"`
         EmailAddress                 string    `json:"email_address,omitempty"`
         HighestEducationalAttainment string    `json:"highest_educational_attainment,omitempty"`
         MotherFirstName              string    `json:"mother_first_name,omitempty"`
         MotherMiddleName             string    `json:"mother_middle_name,omitempty"`
         MotherLastName               string    `json:"mother_last_name,omitempty"`
         ContactNo                    string    `json:"contact_no,omitempty"`
         Address                      string    `json:"address,omitempty"`
     }
     ```
   - **Note:** Do NOT add `DocumentIDs []string` - relationships are derived from documents

2. **Create InhabitantStore Interface**
   - Create `server/core/inhabitant/interfaces.go`:
     ```go
     type InhabitantStore interface {
         Create(ctx context.Context, inh *Inhabitant) (*Inhabitant, error)
         GetByUUID(ctx context.Context, uuid string) (*Inhabitant, error)
         GetByCode(ctx context.Context, code string) (*Inhabitant, error)  // Critical for linking
         Update(ctx context.Context, uuid string, inh *Inhabitant) (*Inhabitant, error)
         Delete(ctx context.Context, uuid string) error
         List(ctx context.Context, offset int, limit int) ([]*Inhabitant, error)
         Search(ctx context.Context, query string) ([]*Inhabitant, error)
         FindByName(ctx context.Context, name string) ([]*Inhabitant, error)  // For migration matching
     }
     
     type CacheStore interface {
         ReloadCache(ctx context.Context) error
     }
     ```

3. **Implement InhabitantFileStore**
   - Create `server/core/inhabitant/store/` folder with files:
     - `store.go` - Main store struct, constructor, initialization
     - `crud.go` - Create, Read, Update, Delete operations
     - `cache.go` - In-memory caching layer
     - `search.go` - Search and FindByName with name→code index
     - `internal.go` - Internal helpers (file I/O, validation)
   - Copy patterns from `server/core/document/store/` but simplify:
     - No `files.json` handling (inhabitants have no attachments)
     - Single `inhabitant.json` per folder
   - Use existing `NamingStrategyPrefixDDDYY("inhabitant")` for folder naming

4. **In-Memory Indexes**
   - `inhabitants map[string]*Inhabitant` (UUID → Inhabitant)
   - `codeToUUID map[string]string` (Code → UUID)
   - `nameIndex map[string][]string` (normalized_name → []UUID) for FindByName

5. **Create SQLite Migration Script**
   - Script location: `server/cmd/scripts/migrate_inhabitants.go`
   - Read all inhabitants from SQLite `app.db`
   - Assign new codes sequentially: `inhabitant-001-26`, `inhabitant-002-26`, ...
   - Write to `/data/inhabitants/{folder_name}/inhabitant.json`
   - Generate old_id→new_code mapping file for reference

6. **Feature Flag Integration**
   - Update `server/cmd/server/main.go`:
     ```go
     var useFileInhabitants = flag.Bool("file-inhabitants", false, "Use file-based inhabitant storage")
     
     // In main():
     var inhabitantService *inhabitant.Service
     if *useFileInhabitants {
         inhabPath := filepath.Join(*dataPath, "inhabitants")
         namingStrategy := document.NewNamingStrategyPrefixDDDYY("inhabitant")
         inhabitantStore, err := inhabitantstore.New(inhabPath, namingStrategy)
         inhabitantService = inhabitant.NewServiceWithStore(inhabitantStore)
     } else {
         inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
         inhabitantService = inhabitant.NewService(inhabitantRepo)
     }
     ```

### File Structure Example
```
/data/inhabitants/inhabitant-001-26/inhabitant.json
```

```json
{
  "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "code": "001-26",
  "folder_name": "inhabitant-001-26",
  "first_name": "John",
  "last_name": "Doe",
  "middle_name": "Dabba",
  "birthdate": "1990-05-15T00:00:00Z",
  "address": "123 Main St, Barangay Example",
  "created_at": "2026-02-05T10:30:00Z",
  "updated_at": "2026-02-05T10:30:00Z"
}
```

---

## Phase 3: Migration & Cleanup

**Goal:** Complete the transition and remove SQLite dependency for inhabitants.

### Tasks

1. **Validate File-Based Store**
   - Run with `--file-inhabitants` flag in staging
   - Compare query results between SQLite and file-based stores
   - Monitor cache warmup time and memory usage

2. **Dual-Write Period (Optional)**
   - Write to both SQLite and file-system
   - Read from file-system only
   - Compare results in background job
   - Alert on mismatches

3. **Remove SQLite Inhabitant Dependency**
   - Remove `inhabitant.NewSQLRepository(db.AppDB)` from main.go
   - Remove `inhabitant.InitSchema(db.AppDB)` call
   - Delete `server/core/inhabitant/store.go` (SQLite implementation)
   - Delete `server/core/inhabitant/repo_sql.go`
   - Update `repo.go` interface to match new `InhabitantStore`
   - Keep `db.AuthDB` for users/sessions (unchanged)

4. **Update Documentation**
   - Update README.md with new folder structure
   - Document backup/restore procedures
   - Add migration guide for existing deployments

---

## Phase 4: Generalize for Custom Documents (Future)

**Goal:** Extract common patterns and apply to other document types.

### Tasks

1. **Extract BaseFolderStore**
   - Once inhabitant store is stable, extract common patterns:
     - File I/O operations
     - In-memory caching
     - CRUD operations
     - Naming strategy integration
   - Location: `server/core/storage/folder_store.go`
   - Use Go generics: `FolderStore[T any]`

2. **Refactor Inhabitant Store**
   - Replace duplicated code with `BaseFolderStore[*Inhabitant]`
   - Maintain same interface

3. **Apply to Custom Folder**
   - Create `/data/custom/` document type
   - Use `FolderStore` with `NewNamingStrategyPrefixDDDYY("custom")`

4. **Unified Search Aggregator**
   - Extend to search across all entity types
   - Filter by entity type in search results

---

## Existing Code Analysis

### Files to Keep (Unchanged)
- `server/core/inhabitant/model.go` - Extend with new fields
- `server/core/inhabitant/repo.go` - Keep interface, update methods
- `server/core/inhabitant/service.go` - Update to support both stores
- `server/core/document/store/` - Pattern to replicate

### Files to Modify
| File | Changes |
|------|---------|
| `server/cmd/server/main.go` | Feature flag, new base paths |
| `server/core/document/model.go` | Add `ParticipantIDs` field |
| `server/core/inhabitant/model.go` | Add UUID, Code, FolderName, timestamps |
| `server/core/search/aggregator.go` | ID-based search, reverse index |

### Files to Create
| File | Purpose |
|------|---------|
| `server/core/inhabitant/interfaces.go` | New InhabitantStore interface |
| `server/core/inhabitant/store/store.go` | Main file-based store |
| `server/core/inhabitant/store/crud.go` | CRUD operations |
| `server/core/inhabitant/store/cache.go` | In-memory caching |
| `server/core/inhabitant/store/search.go` | Search with name index |
| `server/core/inhabitant/store/internal.go` | Internal helpers |
| `server/cmd/scripts/restructure_folders.go` | Folder migration |
| `server/cmd/scripts/migrate_inhabitants.go` | SQLite → file migration |
| `server/cmd/scripts/link_participants.go` | Name→ID matching |

### Files to Delete (Phase 3)
- `server/core/inhabitant/store.go` - SQLite Store implementation
- `server/core/inhabitant/repo_sql.go` - SQL Repository
- `server/core/inhabitant/store_test.go` - SQLite tests

---

## Architectural Decisions

### 1. Grouped Folder Structure
**Decision:** Use grouped folders (`/data/cases/`, `/data/inhabitants/`) from the start

**Rationale:**
- High breaking change tolerance (development phase, not production)
- Cleaner organization for multiple entity types
- Easier backup/restore per entity type
- Future-proof for additional document types

### 2. Documents as Single Source of Truth
**Decision:** No bidirectional sync; derive inhabitant→document relationships from document `participant_ids`

**Rationale:**
- Avoids circular update dependencies
- Prevents race conditions on concurrent updates
- Eliminates orphan reference problems
- Simpler debugging - documents always correct

### 3. Copy-Then-Refactor
**Decision:** Copy document store pattern for inhabitants, extract generics later

**Rationale:**
- Document store is proven and stable
- Two working implementations reveal true common patterns
- Avoids premature abstraction
- Can validate independently before refactoring

### 4. Participant ID Format
**Decision:** Use full code `inhabitant-001-26` in `participant_ids`

**Rationale:**
- Self-describing (prefix identifies entity type)
- No lookup table needed
- Future-proof for other entity types (`organization-001-26`)
- Direct navigation from ID to folder path

### 5. Reuse Existing Naming Strategy
**Decision:** Use `NewNamingStrategyPrefixDDDYY("inhabitant")` from document package

**Rationale:**
- Already tested and working
- Consistent code format across entity types
- No new code needed for naming logic

---

## Success Criteria

### Phase 0 ✅
- [x] All existing case folders moved to `/data/cases/`
- [x] Document store works with new base path
- [x] Empty `/data/inhabitants/` and `/data/custom/` created
- [x] All existing tests pass

### Phase 1
- [ ] Documents can reference inhabitants by ID via `participant_ids`
- [ ] Name→ID matching script generates accurate mappings
- [ ] Reverse index (inhabitant → documents) built from document data
- [ ] API endpoint returns documents for a given inhabitant code
- [ ] Backward compatibility maintained for documents without `participant_ids`

### Phase 2
- [ ] InhabitantStore interface implemented with all required methods
- [ ] Inhabitants stored in `/data/inhabitants/inhabitant-XXX-YY/inhabitant.json`
- [ ] In-memory cache with codeToUUID and nameIndex
- [ ] Migration script exports all SQLite data without loss
- [ ] Feature flag allows switching between SQLite and file stores

### Phase 3
- [ ] SQLite dependency removed from inhabitant storage
- [ ] Auth DB remains functional (unchanged)
- [ ] Zero data loss validated
- [ ] Documentation updated

---

## Implementation Timeline

**Phase 0 (Folder Restructuring):** 2-3 days
- Migration script: 1 day
- Update document store paths: 0.5 days
- Testing: 1 day

**Phase 1 (Document-Inhabitant Linking):** 1-2 weeks
- Schema updates: 1 day
- Name matching script: 3-4 days
- Reverse index: 2 days
- API endpoint: 1 day
- Testing: 2 days

**Phase 2 (Inhabitant File Store):** 2-3 weeks
- Model updates: 1 day
- Interface definition: 0.5 days
- Store implementation: 1 week
- Migration script: 2 days
- Feature flag integration: 1 day
- Testing: 3 days

**Phase 3 (Cleanup):** 1 week
- Validation period: 3-4 days
- Code removal: 1 day
- Documentation: 2 days

**Total for Phases 0-3:** 5-7 weeks

---

## Risk Mitigation

### 1. Folder Restructure Breaks Paths
**Risk:** Existing code references old paths

**Mitigation:**
- Update all path references in single commit
- Run full test suite after migration
- Keep backup of original `/data` folder

### 2. Name Matching Errors
**Risk:** Incorrect participant→inhabitant mappings

**Mitigation:**
- Generate report mode for manual review
- Confidence scores for each match
- Keep original name fields for fallback
- Provide admin UI to correct mismatches

### 3. Migration Data Loss
**Risk:** SQLite data not fully exported

**Mitigation:**
- Validation step comparing record counts
- Feature flag allows instant rollback
- Keep SQLite database until Phase 3 complete

### 4. Performance Degradation
**Risk:** File-based queries slower than SQLite

**Mitigation:**
- Aggressive in-memory caching
- Index building during startup
- Benchmark against SQLite before cutover
