package spiceworks

import (
	"encoding/json"
	"html"
	"io/ioutil"
)

type Ticket struct {
	Id       int
	Summary  string
	Assignee struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
}

// Tickets makes a HTTP request to the running Spiceworks instance and returns
// a slice of Tickets. If filter is the empty string (""), this method returns
// the first 100 tickets (currently, this method doesn't handle requesting more
// than a single 'page' of tickets, so the limit is 100). If filter is
// specified, only the tickets which pass the filter condition are returned.
//
// Currently, the only known filter is "open".
func (c *Client) Tickets(filter string) (tickets []*Ticket, err error) {
	var ticketsUrl string
	if filter == "" {
		ticketsUrl = c.BaseUrl + "api/tickets.json"
	} else {
		ticketsUrl = c.BaseUrl + "api/tickets.json?filter=" + filter
	}

	resp, err := c.HttpClient.Get(ticketsUrl)
	if err != nil {
		return tickets, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tickets, err
	}

	if err = json.Unmarshal(body, &tickets); err != nil {
		return tickets, err
	}

	for i, ticket := range tickets {
		tickets[i].Summary = html.UnescapeString(ticket.Summary)
	}

	return tickets, nil
}
