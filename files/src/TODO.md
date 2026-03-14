# TODO & Unimplemented Items

## Summary

| Area | Status | Items Remaining |
|------|--------|-----------------|
| Progress Reports | ✅ Implemented | — |
| Comments | ✅ Implemented | — |
| Health Check | ✅ Implemented | SES check (optional) |
| Team Deletion | ✅ Implemented | S3 file cleanup |
| Invite Expiration | Partial | Missing scheduler trigger |

---

## Partial Implementations

### Team Deletion — S3 File Cleanup

**File:** `db/teams.go` — `DeleteTeamByID`

DynamoDB cascade delete is fully implemented (members, invites, seasons, goals, progress reports, comments, comment files). The following S3 objects are **not** deleted:

- [ ] Team picture (`teams/{teamId}/...`)
- [ ] Goal pictures (`goals/{goalId}/...`)
- [ ] Comment files (`comments/{commentId}/...`)

S3 cleanup would require storing the S3 key on each record and calling `s3.DeleteObject` during the cascade, or using an S3 lifecycle policy.

---

### Invite Expiration — Missing Trigger

**File:** `db/invites.go` — line ~318

`ExpireInvites()` is implemented as a standalone DB function but is never called. Expired invites remain in `pending` status indefinitely.

- [ ] Create an EventBridge scheduled rule (e.g., daily) to trigger a Lambda that calls `ExpireInvites()`
- [ ] Alternatively, check expiry lazily on invite fetch and update status at read time

---

## Minor Notes

- `TeamSettings.AllowFileUploads` is stored and returned but **not enforced** in the file upload presign endpoints — the flag should gate access to `/picture/presign` and `/file/presign` routes.
- No rate limiting is implemented on public endpoints (`/invites/complete`, `/invites/:inviteToken`).
- The `Birthdate` field on `User` is stored but never validated (format, plausibility).
- SES health check is not included in `POST /health` — only DynamoDB, S3, and Cognito are checked.
