package model

import (
	"fmt"
	"strconv"
	"strings"
)

type ListQuery struct {
	Page         int `query:"page" json:"page"`
	PagePerCount int `query:"page_per_count" json:"page_per_count"`

	ExcludeDeleted bool    `query:"exclude_deleted" json:"exclude_deleted"`   // default false
	PlayerUUID     *string `query:"player_uuid" json:"player_uuid,omitempty"` // filter by player
	EquipmentID    []uint  `query:"eq_id" json:"eq_id,omitempty"`             // filter by eq ids
	ItemID         []uint  `query:"item_id" json:"item_id,omitempty"`         // filter by item ids
	Sorts          *string `query:"sort" json:"sort_by,omitempty"`            // comma‑separated field list
	Orders         *string `query:"order" json:"order_by,omitempty"`          // matching order list (asc/desc)

	Search *string `query:"q" json:"q,omitempty"` // matching id or name
}

func (q *ListQuery) ShouldSearch() (bool, string) {
	if q.Search != nil && len(*q.Search) > 0 {
		return true, *q.Search
	}
	return false, ""
}

func (q *ListQuery) SearchAsID() (int, error) {
	//parse as int, if fail return 0 and error
	return strconv.Atoi(*q.Search)
}

func (q *ListQuery) Generalize() {
	if q.Page <= 0 {
		q.Page = 1
	}

	if q.PagePerCount <= 0 {
		q.PagePerCount = 10
	}

	// if q.PagePerCount > 10000 {
	// 	q.PagePerCount = 10000
	// }
}

func (q *ListQuery) Offset() int {
	if q.Page <= 1 {
		return 0
	}
	return (q.Page - 1) * q.PagePerCount
}

func (q *ListQuery) IsForSpecificPlayer() (bool, SessionBase) {
	if q.PlayerUUID != nil && len(*q.PlayerUUID) > 0 {
		return true, FormSessionBase(*q.PlayerUUID)
	}
	return false, SessionBase{}
}

func (q *ListQuery) IsForSpecificEquipmentIDs() (bool, []uint) {
	if q.EquipmentID != nil && len(q.EquipmentID) > 0 {
		return true, q.EquipmentID
	}
	return false, nil
}

func (q *ListQuery) IsForSpecificItemIDs() (bool, []uint) {
	if q.ItemID != nil && len(q.ItemID) > 0 {
		return true, q.ItemID
	}
	return false, nil
}

func (q *ListQuery) GeneralizeSortOrder() {
	// Ensure Orders is set to a valid value (asc or desc). Default to desc.
	defaultOrder := "desc"
	if q.Orders == nil {
		q.Orders = &defaultOrder
		return
	}

	lo := strings.ToLower(strings.TrimSpace(*q.Orders))
	if lo != "asc" && lo != "desc" {
		q.Orders = &defaultOrder
	}
}

func (q *ListQuery) ShouldSortBy() (res string, do bool) {
	if q.Sorts != nil && len(*q.Sorts) > 0 {
		q.GeneralizeSortOrder()
		do = true
		res = fmt.Sprintf("%s %s", *q.Sorts, *q.Orders)
		return
	}
	do = false
	return
}

type PaginationMeta struct {
	ListQuery
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

func (q *PaginationMeta) Generalize() {
	q.ListQuery.Generalize()

	if q.TotalCount > 0 {
		q.TotalPages = int(q.TotalCount / int64(q.PagePerCount))
		if q.TotalCount%int64(q.PagePerCount) != 0 {
			q.TotalPages++
		}
	} else {
		q.TotalPages = 0
	}
}
