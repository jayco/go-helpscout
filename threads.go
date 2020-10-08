package helpscout

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ThreadLister ..
type ThreadLister interface {
	Process(thread Thread) bool
}

const (
	// ThreadTypeBeaconchat ..
	ThreadTypeBeaconchat = "beaconchat"

	// ThreadTypeChat ..
	ThreadTypeChat = "chat"

	// ThreadTypeCustomer ..
	ThreadTypeCustomer = "customer"

	// ThreadTypeForwardChild ..
	ThreadTypeForwardChild = "forwardchild"

	// ThreadTypeForwardParent ..
	ThreadTypeForwardParent = "forwardparent"

	// ThreadTypeLineitem ..
	ThreadTypeLineitem = "lineitem"

	// ThreadTypeMessage ..
	ThreadTypeMessage = "message"

	// ThreadTypeNote ..
	ThreadTypeNote = "note"

	// ThreadTypePhone ..
	ThreadTypePhone = "phone"

	// ThreadTypeReply ..
	ThreadTypeReply = "reply"

	// ThreadStatusActive ..
	ThreadStatusActive = "active"

	// ThreadStatusClosed ..
	ThreadStatusClosed = "closed"

	// ThreadStatusNochange ..
	ThreadStatusNochange = "nochange"

	// ThreadStatusPending ..
	ThreadStatusPending = "pending"

	// ThreadStatusSpam ..
	ThreadStatusSpam = "spam"

	// ThreadStateDraft ..
	ThreadStateDraft = "draft"

	// ThreadStateHidden ..d
	ThreadStateHidden = "hidden"

	// ThreadStatePublished ..
	ThreadStatePublished = "published"

	// ThreadStateReview ..
	ThreadStateReview = "review"
)

// ThreadCreator ..
type ThreadCreator struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	FirstName string `json:"first"`
	LastName  string `json:"last"`
	Email     string `json:"email"`
}

// ThreadSource ..
type ThreadSource struct {
	Via  string `json:"via"`
	Type string `json:"type"`
}

// Thread ..
type Thread struct {
	ID           int           `json:"id"`
	Type         string        `json:"type"`
	AssignedTo   User          `json:"assignedTo"`
	Status       string        `json:"status"`
	State        string        `json:"state"`
	Body         string        `json:"body"`
	Source       ThreadSource  `json:"source"`
	Customer     Customer      `json:"customer"`
	CreatedBy    ThreadCreator `json:"createdBy"`
	SavedReplyID int           `json:"savedReplyId"`
	To           []string      `json:"to"`
	CC           []string      `json:"cc"`
	BCC          []string      `json:"bcc"`
	CreatedAt    time.Time     `json:"createdAt"`
	OpenedAt     time.Time     `json:"openedAt"`
}

// ListThreads ..
func (c *Client) ListThreads(conversationID int, lister ThreadLister) error {
	resource := fmt.Sprintf("/conversations/%d/threads", conversationID)

	query := &url.Values{}
	page := 1
	for {
		var tList struct {
			Threads []Thread `json:"threads"`
		}

		req := &generalListAPICallReq{
			Embedded: &tList,
		}

		err := c.doAPICall(http.MethodGet, resource, query, nil, req)
		if err != nil {
			return err
		}

		if req.Page.TotalPages == 0 {
			break
		}

		for _, thread := range tList.Threads {
			if !lister.Process(thread) {
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
