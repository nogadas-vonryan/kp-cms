- GET /
  - Public
  - Response 200 (text/plain): "Archivist Server - Prototype"

- GET /health
  - Public
  - Response 200 (text/plain): "ok"

- GET /documents
  - Auth required (AuthMiddleware)
  - Query params: offset (int, default 0), limit (int, default 15)
  - Response 200 application/json: array of Document
  - Sample:
````json
[
  {
    "uuid": "6f1b2c3d-...-abcd",
    "code": "DOC-001",
    "folder_name": "DOC-001-Example Title",
    "title": "Example Title",
    "fields": { "author": "Alice", "category": "report" },
    "files": [
      { "file_name": "scan.pdf", "type": ".pdf", "size": 12345, "created_at": "2025-12-01T12:34:56Z" }
    ],
    "created_at": "2025-12-01T12:00:00Z",
    "updated_at": "2025-12-01T12:34:56Z"
  }
]
````

- GET /documents/{uuid}
  - Auth required
  - Response 200 application/json: single Document (same shape as above)
  - 404 example:
````json
{ "error": "document not found" }
````

- GET /documents/code/{code}
  - Auth required
  - Response 200 application/json: single Document
  - 404 same error payload as above

- POST /documents
  - Admin only (AuthMiddleware + RequireRole(RoleAdmin))
  - Request application/json (CreateDocumentRequest):
````json
{
  "code": "OPTIONAL_OR_EMPTY",
  "folder_name": "OPTIONAL_OR_EMPTY",
  "title": "New Document Title",
  "fields": { "author": "Bob" }
}
````
  - Response 201 application/json: created Document

- PUT /documents/{uuid}
  - Admin only
  - Request application/json (UpdateDocumentRequest):
````json
{
  "code": "DOC-002",
  "title": "Updated Title",
  "fields": { "status": "archived" }
}
````
  - Response 200 application/json: updated Document
  - 404 uses { "error": "document not found" }

- POST /documents/{uuid}/files
  - Admin only
  - Request: multipart/form-data; field name "file" (file upload, max ~32MB parsed)
  - Response 201 application/json:
````json
{ "message": "file uploaded successfully", "file_name": "scan.pdf" }
````

- DELETE /documents/{uuid}
  - Admin only
  - Response 204 No Content
  - 404 uses { "error": "document not found" }

- GET /documents/conflicts
  - Admin only
  - Response 200 application/json: array of SyncIssue
  - Sample:
````json
[
  { "type": "DUPLICATE_CODE", "path": "0001-old", "message": "code 'DOC-001' already claimed by folder '0001-new'" }
]
````

- POST /documents/reload
  - Admin only
  - Response 200 application/json:
````json
{
  "status": "success",
  "conflicts": [
    { "type": "MISSING_META", "path": "orphan-folder", "message": "could not read meta.json: ..." }
  ]
}
````

Notes:
- Common error response shape: { "error": "message" } (used for validation / not found / server errors).
- The handler for deleting an individual file exists in code but is not registered in routes; currently only file upload is routed.
- All /documents endpoints require the top-level AuthMiddleware; admin endpoints additionally require RoleAdmin.