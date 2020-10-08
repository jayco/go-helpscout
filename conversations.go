package helpscout

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ConversationLister ..
type ConversationLister interface {
	Process(c Conversation) bool
}

// ConditionType ..
type ConditionType int

const (
	// Inclusively ..
	Inclusively ConditionType = 1

	// Exclusively ..
	Exclusively ConditionType = 2

	// ByUser ..
	ByUser = "user"

	// ByCustomer ..
	ByCustomer = "customer"

	// ConversationTypeEmail ..
	ConversationTypeEmail = "email"

	// ConversationTypeChat ..
	ConversationTypeChat = "chat"

	// ConversationTypePhone ..
	ConversationTypePhone = "phone"

	// ConversationStatusOpen ..
	ConversationStatusOpen = "open"

	// ConversationStatusClosed ..
	ConversationStatusClosed = "closed"

	// ConversationStatusActive ..
	ConversationStatusActive = "active"

	// ConversationStatusPending ..
	ConversationStatusPending = "pending"

	// ConversationStatusSpam ..
	ConversationStatusSpam = "spam"

	// ConversationStatePublished ..
	ConversationStatePublished = "published"

	// ConversationStateDraft ..
	ConversationStateDraft = "draft"

	// ConversationStateDeleted  ..
	ConversationStateDeleted = "deleted"
)

type filterIntValues struct {
	cType  ConditionType
	values []int
}

func (v *filterIntValues) Set(values []int, cType ConditionType) {
	v.values = values
	v.cType = cType
}

type filterStringValues struct {
	cType  ConditionType
	values []string
}

func (v *filterStringValues) Set(values []string, cType ConditionType) {
	v.values = values
	v.cType = cType
}

type filterTimePeriod struct {
	cType ConditionType
	from  time.Time
	to    time.Time
}

func (v *filterTimePeriod) Set(from time.Time, to time.Time, cType ConditionType) {
	v.from = from
	v.to = to
	v.cType = cType
}

// ConversationLookupFilter ..
type ConversationLookupFilter struct {
	mailboxIds    *filterIntValues
	statuses      *filterStringValues
	types         *filterStringValues
	states        *filterStringValues
	createdPeriod *filterTimePeriod
	updatedPeriod *filterTimePeriod
}

// NewConversationLookupFilter ..
func NewConversationLookupFilter() *ConversationLookupFilter {
	return &ConversationLookupFilter{}
}

func getConditionType(cType []ConditionType) ConditionType {
	if len(cType) == 0 {
		return Inclusively
	}

	if len(cType) > 1 {
		panic("There must be only one condition type")
	}

	return cType[0]
}

// MailboxIds ..
func (f *ConversationLookupFilter) MailboxIds(ids []int, cType ...ConditionType) {
	if f.mailboxIds == nil {
		f.mailboxIds = &filterIntValues{}
	}
	f.mailboxIds.Set(ids, getConditionType(cType))
}

// Status ..
func (f *ConversationLookupFilter) Status(statuses []string, cType ...ConditionType) {
	if f.statuses == nil {
		f.statuses = &filterStringValues{}
	}
	f.statuses.Set(statuses, getConditionType(cType))
}

// State ..
func (f *ConversationLookupFilter) State(states []string, cType ...ConditionType) {
	if f.states == nil {
		f.states = &filterStringValues{}
	}
	f.states.Set(states, getConditionType(cType))
}

// Type ..
func (f *ConversationLookupFilter) Type(types []string, cType ...ConditionType) {
	if f.types == nil {
		f.types = &filterStringValues{}
	}
	f.types.Set(types, getConditionType(cType))
}

// CreatedTime ..
func (f *ConversationLookupFilter) CreatedTime(from time.Time, to time.Time, cType ...ConditionType) {
	if f.createdPeriod == nil {
		f.createdPeriod = &filterTimePeriod{}
	}
	f.createdPeriod.Set(from, to, getConditionType(cType))
}

// ModifiedTime ..
func (f *ConversationLookupFilter) ModifiedTime(from time.Time, to time.Time, cType ...ConditionType) {
	if f.updatedPeriod == nil {
		f.updatedPeriod = &filterTimePeriod{}
	}
	f.updatedPeriod.Set(from, to, getConditionType(cType))
}

// AnsweredBy ..
type AnsweredBy struct {
	Time               time.Time `json:"time"`
	FriendlyWaitPeriod string    `json:"friendly"`
	By                 string    `json:"latestReplyFrom"`
}

// ConversationSource ..
type ConversationSource struct {
	Via  string `json:"via"`
	Type string `json:"type"`
}

// ConversationCustomer ..
type ConversationCustomer struct {
	ID int `json:"id"`
}

// Conversation ..
type Conversation struct {
	ID              int                  `json:"id"`
	Number          int                  `json:"number"`
	Threads         int                  `json:"threads"`
	Type            string               `json:"type"`
	FolderID        int                  `json:"folderId"`
	Status          string               `json:"status"`
	State           string               `json:"state"`
	Subject         string               `json:"subject"`
	Preview         string               `json:"preview"`
	MailboxID       int                  `json:"mailboxId"`
	Assignee        User                 `json:"assignee"`
	CreatedBy       User                 `json:"createdBy"`
	CreatedAt       time.Time            `json:"createdAt"`
	ClosedAt        time.Time            `json:"closedAt"`
	UpdatedAt       time.Time            `json:"userUpdatedAt"`
	ClosedBy        int                  `json:"closedBy"`
	Answered        AnsweredBy           `json:"customerWaitingSince"`
	Source          ConversationSource   `json:"source"`
	Tags            []TagShort           `json:"tags"`
	CC              []string             `json:"cc"`
	BCC             []string             `json:"bcc"`
	PrimaryCustomer ConversationCustomer `json:"primaryCustomer"`
	CustomFields    []CustomField        `json:"customFields"`
}

// ListConversations ..
func (c *Client) ListConversations(filter *ConversationLookupFilter, lister ConversationLister) error {
	query, err := prepareListConversationQuery(filter)
	if err != nil {
		return err
	}

	if filter.statuses == nil {
		return c.listConversationsImpl(query, lister)
	}

	statuses := prepareListOfStatuses(filter)
	for _, status := range statuses {
		query.Set("status", status)
		err := c.listConversationsImpl(query, lister)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) listConversationsImpl(query *url.Values, lister ConversationLister) error {
	page := 1
	query.Del("page")
	for {
		var cList struct {
			Conversations []Conversation `json:"conversations"`
		}

		req := &generalListAPICallReq{
			Embedded: &cList,
		}
		err := c.doAPICall(http.MethodGet, "/conversations", query, nil, req)
		if err != nil {
			return err
		}

		if req.Page.TotalPages == 0 {
			break
		}

		for _, conversation := range cList.Conversations {
			if !lister.Process(conversation) {
				return ErrorInterrupted
			}
		}

		if req.Page.Number == req.Page.TotalPages {
			break
		}

		page++
		query.Set("page", strconv.Itoa(page))
	}

	return nil
}

func prepareListOfStatuses(filter *ConversationLookupFilter) []string {
	var statuses []string
	if filter.statuses != nil {
		switch filter.statuses.cType {
		case Inclusively:
			statuses = append(statuses, filter.statuses.values...)
		case Exclusively:
			m := map[string]int{
				ConversationStatusOpen:    0,
				ConversationStatusClosed:  0,
				ConversationStatusActive:  0,
				ConversationStatusPending: 0,
				ConversationStatusSpam:    0,
			}

			for _, v := range filter.statuses.values {
				delete(m, v)
			}

			for k := range m {
				statuses = append(statuses, k)
			}
		default:
			panic("Unknown condition type")
		}
	}

	return statuses
}

func prepareListConversationQuery(filter *ConversationLookupFilter) (*url.Values, error) {
	if filter == nil {
		return &url.Values{}, nil
	}

	queryValues := []string{}

	if filter.mailboxIds != nil {
		b := make([]string, len(filter.mailboxIds.values))
		for i, v := range filter.mailboxIds.values {
			b[i] = fmt.Sprintf("mailboxid:%d", v)
		}

		queryValues = append(queryValues, fmt.Sprintf("(%s)", strings.Join(b, " OR ")))
	}

	if filter.createdPeriod != nil {
		fromStr, toStr := formatFromToTimePeriod(filter.createdPeriod.from, filter.createdPeriod.to)
		queryValues = append(queryValues, fmt.Sprintf("createdAt:[%s TO %s]", fromStr, toStr))
	}

	if filter.updatedPeriod != nil {
		fromStr, toStr := formatFromToTimePeriod(filter.updatedPeriod.from, filter.updatedPeriod.to)
		queryValues = append(queryValues, fmt.Sprintf("modifiedAt:[%s TO %s]", fromStr, toStr))
	}

	query := url.Values{}
	if len(queryValues) != 0 {
		query.Set("query", fmt.Sprintf("(%s)", strings.Join(queryValues, " AND ")))
	}

	return &query, nil
}

func formatFromToTimePeriod(from time.Time, to time.Time) (string, string) {
	fromStr := "*"
	if !from.IsZero() {
		fromStr = from.Format("2006-01-02T15:04:05Z")
	}

	toStr := "*"
	if !to.IsZero() {
		toStr = to.Format("2006-01-02T15:04:05Z")
	}

	return fromStr, toStr
}
