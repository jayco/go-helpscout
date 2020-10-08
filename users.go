package helpscout

import (
	"net/http"
	"net/url"
	"strconv"
)

// UsersLister ..
type UsersLister interface {
	Process(c User) bool
}

// User ..
type User struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	FirstName string `json:"first"`
	LastName  string `json:"last"`
	Email     string `json:"email"`
}

// ListUsers ..
func (c *Client) ListUsers(lister UsersLister) error {
	page := 1
	query := &url.Values{}
	for {
		var uList struct {
			Users []User `json:"users"`
		}

		req := &generalListAPICallReq{
			Embedded: &uList,
		}
		err := c.doAPICall(http.MethodGet, "/users", query, nil, req)
		if err != nil {
			return err
		}

		if req.Page.TotalPages == 0 {
			break
		}

		for _, user := range uList.Users {
			if !lister.Process(user) {
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
