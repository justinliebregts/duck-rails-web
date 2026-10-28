package errors

import "errors"

var (
	NoAuth                = errors.New("no auth header")
	NoAdminPrivilegeFound = errors.New("no admin privilege found")

	ErrPrivilegeExpired = errors.New("privilege expired")

	ErrSelfPrivilegePathMissing = errors.New("SelfPrivilege: Path Not Found")
	ErrSelfPrivilegeParseFailed = errors.New("SelfPrivilege: Parse Failed")
	ErrSelfPrivilegeNotOwner    = errors.New("SelfPrivilege: ResourceNotOwned")

	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header")

	ErrMetaQuestAttestationFailed = errors.New("MetaQuestAttestation: Failed")
	ErrMetaQuestAttestationBanned = errors.New("MetaQuestAttestation: Banned")
	ErrMetaQuestAttestationMissed = errors.New("MetaQuestAttestation: Missed")

	ErrAppleInvalidSignature = errors.New("AppleInvalidSignature")

	ErrInvalidTimestamp = errors.New("invalid timestamp")
	ErrTimestampTooOld  = errors.New("auth timestamp is expired. It is older than the allowed MaxAge window")
	ErrTimestampTooNew  = errors.New("auth timestamp is from the future. Check your device's clock synchronization")
)
