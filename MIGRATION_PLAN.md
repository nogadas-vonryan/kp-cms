# Migration Plan: Metadata Storage Overhaul

Move document metadata from per-folder JSON files into a single SQLite DB at {dataPath}/data.db while keeping attachments on disk, and keep auth.db unchanged. Use a gradual, backward-compatible migration where reads consult data.db first and fall back to existing JSON, then progressively migrate and remove JSON writes.

## Phases

### Phase 1: Baseline Inventory and Configuration
- Document current storage flows (current state):
  - `server/core/document/store/docs.go` writes `meta.json` and initializes `files.json` on create.
  - `server/core/document/store/io.go` reads `files.json`, syncs on-disk files, and updates metadata.
  - `server/core/document/store/cache.go` rebuilds cache from `meta.json` and auto-heals code mismatches.
- Confirm auth remains in auth.db and is not affected:
  - `server/core/database/database.go` determines the platform config directory for auth.db.
  - `server/core/auth/` uses auth.db for users/sessions/RBAC.
- Confirm dataPath and backupPath wiring:
  - `main.go` and `server/cmd/server/main.go` default to `./data` and `./backup` and pass them to `store.New`.
  - `app.go` configuration persists dataPath and backupPath in the Wails app config.

### Phase 1 Verification
- Confirm a new document still creates a folder with `meta.json` and `files.json`.
- Confirm file upload/download still works for one representative document.
- Confirm the cache can reload existing documents from `meta.json` after a restart.
- Confirm auth login/session behavior still works against auth.db.

### Phase 2: Define DB Schema and Access Layer
- Create a new SQLite-backed metadata store in `server/core/document/store/`.
- Proposed tables:
  - `documents`: uuid, code, folder_name, title, fields_json, created_at, updated_at
  - `files`: id, document_uuid, filename, description, note, tags_json, type, size, created_at
- Add indexes on uuid, code, and folder_name.
- Status: implemented in `server/core/document/store/metadata_sql.go` (no existing metadata DB layer was found).

### Phase 3: Dual-Read and Dual-Write
- Reads: check data.db first; if missing, read legacy meta.json/files.json.
- Writes: continue writing JSON while also writing to data.db.
- Goal: safe, reversible rollout.

### Phase 4: Migration Tooling
- Add a migration utility that scans data folders, loads meta.json/files.json, and populates data.db.
- Provide a CLI flag or admin endpoint to run migration in batches.
- Record migrated folders for progress tracking.

### Phase 5: Attachment Layout Enforcement
- Enforce attachments layout:
  - `{dataPath}/cases/case-XXX-YY/attachments/*`
  - `{dataPath}/inhabitants/inhabitant-XXX-YY/attachments/*`
- Remove metadata JSON from attachment folders in new writes.
- Update upload/download/delete handlers to use data.db for metadata.

### Phase 6: DB-Only Writes
- Switch to DB-only writes behind a feature flag (e.g., `metadataStore = db`).
- Keep legacy JSON reads for safety.

### Phase 7: Final Cutover
- Disable legacy reads.
- Remove JSON loaders and add an optional cleanup tool to delete old JSON files.

## Recommended Rollout Order
1. Dual-read and dual-write (Phase 3) in a staging environment.
2. Run migration tooling and validate coverage.
3. Switch to DB-only writes.
4. After validation, disable legacy reads and delete old JSON files.

## Verification Checklist
1. Migrate a small dataset and validate data.db rows match legacy JSON.
2. Upload/download/delete files after migration and confirm DB rows update.
3. Run document and auth API tests to detect regressions.
4. Restart server and ensure cache rebuilds from data.db.

## Decisions
- auth.db remains unchanged; no sessions.db is introduced.
- data.db location: {dataPath}/data.db.
- Attachment folder layout should match the requested structure exactly.
