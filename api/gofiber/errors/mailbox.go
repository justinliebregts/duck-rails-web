package errors

import "errors"

var (
	ErrMailboxEmpty              = errors.New("mailbox_empty")
	ErrMailboxTemplateNotFound   = errors.New("mailbox_template_not_found")
	ErrMailboxInvalidPlayer      = errors.New("mailbox_invalid_player")
	ErrMailboxInvalidEquipmentID = errors.New("mailbox_invalid_equipment_id")
	ErrMailboxInvalidGemAmount   = errors.New("mailbox_invalid_gem_amount")
	ErrMailboxInvalidSend        = errors.New("mailbox_invalid_send")
	ErrMailboxClaimFailed        = errors.New("mailbox_claim_failed")
	ErrMailboxServiceUnavailable = errors.New("mailbox_service_unavailable")
	ErrMailboxItemNotFound       = errors.New("mailbox_item_not_found")
	ErrMailboxItemNotPending     = errors.New("mailbox_item_not_pending")
)
