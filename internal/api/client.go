package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the osTicket API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new osTicket API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request represents the API request body
type Request struct {
	Query      string                 `json:"query"`
	Condition  string                 `json:"condition"`
	Sort       string                 `json:"sort,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Response represents the API response
type Response struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Time    float64         `json:"time,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// TicketData represents ticket response data
type TicketData struct {
	Total   int        `json:"total"`
	Tickets [][]Ticket `json:"tickets"`
}

// Ticket represents a single ticket
type Ticket struct {
	TicketID    int    `json:"ticket_id"`
	TicketPID   int    `json:"ticket_pid"`
	Number      string `json:"number"`
	UserID      int    `json:"user_id"`
	UserEmailID int    `json:"user_email_id"`
	StatusID    int    `json:"status_id"`
	DeptID      int    `json:"dept_id"`
	SLAID       int    `json:"sla_id"`
	TopicID     int    `json:"topic_id"`
	StaffID     int    `json:"staff_id"`
	TeamID      int    `json:"team_id"`
	EmailID     int    `json:"email_id"`
	LockID      int    `json:"lock_id"`
	Flags       int    `json:"flags"`
	Sort        int    `json:"sort"`
	Subject     string `json:"subject"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	IPAddress   string `json:"ip_address"`
	Source      string `json:"source"`
	SourceExtra string `json:"source_extra"`
	IsOverdue   int    `json:"isoverdue"`
	IsAnswered  int    `json:"isanswered"`
	DueDate     string `json:"duedate"`
	EstDueDate  string `json:"est_duedate"`
	Reopened    string `json:"reopened"`
	Closed      string `json:"closed"`
	LastUpdate  string `json:"lastupdate"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

// UserData represents user response data
type UserData struct {
	Total int    `json:"total"`
	Users []User `json:"users"`
}

// User represents a single user
type User struct {
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
	Created string `json:"created"`
}

// DepartmentData represents department response data
type DepartmentData struct {
	Total       int          `json:"total"`
	Departments []Department `json:"departments"`
}

// Department represents a single department
type Department struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TopicData represents topic response data
type TopicData struct {
	Total  int     `json:"total"`
	Topics []Topic `json:"topics"`
}

// Topic represents a single topic
type Topic struct {
	TopicID int    `json:"topic_id"`
	Topic   string `json:"topic"`
}

// SLAData represents SLA response data
type SLAData struct {
	Total int   `json:"total"`
	SLA   []SLA `json:"sla"`
}

// SLA represents a single SLA plan
type SLA struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	GracePeriod int    `json:"grace_period"`
}

// doRequest performs the API request (POST)
func (c *Client) doRequest(req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status == "Error" {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	return &apiResp, nil
}

// doGetRequest performs a GET API request with JSON body
func (c *Client) doGetRequest(req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("GET", c.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status == "Error" {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	return &apiResp, nil
}

// GetTicket gets a specific ticket by ID or number (uses GET)
func (c *Client) GetTicket(id string) (*TicketData, error) {
	resp, err := c.doGetRequest(Request{
		Query:      "ticket",
		Condition:  "specific",
		Parameters: map[string]interface{}{"id": id},
	})
	if err != nil {
		return nil, err
	}

	// Try parsing as TicketData first ([][]Ticket format)
	var data TicketData
	if err := json.Unmarshal(resp.Data, &data); err == nil {
		return &data, nil
	}

	// Try parsing as flat ticket array ([]Ticket format)
	var flatData struct {
		Total   int      `json:"total"`
		Tickets []Ticket `json:"tickets"`
	}
	if err := json.Unmarshal(resp.Data, &flatData); err == nil {
		// Convert flat array to nested format for consistency
		var nestedTickets [][]Ticket
		for _, t := range flatData.Tickets {
			nestedTickets = append(nestedTickets, []Ticket{t})
		}
		return &TicketData{
			Total:   flatData.Total,
			Tickets: nestedTickets,
		}, nil
	}

	// Try parsing as single ticket object
	var singleData struct {
		Total  int    `json:"total"`
		Ticket Ticket `json:"ticket"`
	}
	if err := json.Unmarshal(resp.Data, &singleData); err == nil {
		return &TicketData{
			Total:   singleData.Total,
			Tickets: [][]Ticket{{singleData.Ticket}},
		}, nil
	}

	return nil, fmt.Errorf("failed to parse ticket data: unexpected response format")
}

// GetTicketsByStatus gets tickets by status
func (c *Client) GetTicketsByStatus(status int) (*TicketData, error) {
	resp, err := c.doRequest(Request{
		Query:      "ticket",
		Condition:  "all",
		Sort:       "status",
		Parameters: map[string]interface{}{"status": status},
	})
	if err != nil {
		return nil, err
	}

	var data TicketData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse ticket data: %w", err)
	}

	return &data, nil
}

// GetTicketsByDateRange gets tickets by creation date range
func (c *Client) GetTicketsByDateRange(startDate, endDate string) (*TicketData, error) {
	resp, err := c.doRequest(Request{
		Query:     "ticket",
		Condition: "all",
		Sort:      "creationDate",
		Parameters: map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
	if err != nil {
		return nil, err
	}

	var data TicketData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse ticket data: %w", err)
	}

	return &data, nil
}

// CreateTicketParams contains parameters for creating a ticket
type CreateTicketParams struct {
	Title      string
	Subject    string
	UserID     int
	PriorityID int
	StatusID   int
	DeptID     int
	SLAID      int
	TopicID    int
}

// CreateTicket creates a new ticket
func (c *Client) CreateTicket(params CreateTicketParams) (int, error) {
	resp, err := c.doRequest(Request{
		Query:     "ticket",
		Condition: "add",
		Parameters: map[string]interface{}{
			"title":       params.Title,
			"subject":     params.Subject,
			"user_id":     params.UserID,
			"priority_id": params.PriorityID,
			"status_id":   params.StatusID,
			"dept_id":     params.DeptID,
			"sla_id":      params.SLAID,
			"topic_id":    params.TopicID,
		},
	})
	if err != nil {
		return 0, err
	}

	var ticketID int
	if err := json.Unmarshal(resp.Data, &ticketID); err != nil {
		return 0, fmt.Errorf("failed to parse ticket ID: %w", err)
	}

	return ticketID, nil
}

// ReplyToTicket adds a reply to a ticket
func (c *Client) ReplyToTicket(ticketID int, body string, staffID int) error {
	_, err := c.doRequest(Request{
		Query:     "ticket",
		Condition: "reply",
		Parameters: map[string]interface{}{
			"ticket_id": ticketID,
			"body":      body,
			"staff_id":  staffID,
		},
	})
	return err
}

// CloseTicketParams contains parameters for closing a ticket
type CloseTicketParams struct {
	TicketID int
	Body     string
	StaffID  int
	StatusID int
	TeamID   int
	DeptID   int
	TopicID  int
	Username string
}

// CloseTicket closes a ticket
func (c *Client) CloseTicket(params CloseTicketParams) error {
	_, err := c.doRequest(Request{
		Query:     "ticket",
		Condition: "close",
		Parameters: map[string]interface{}{
			"ticket_id": params.TicketID,
			"body":      params.Body,
			"staff_id":  params.StaffID,
			"status_id": params.StatusID,
			"team_id":   params.TeamID,
			"dept_id":   params.DeptID,
			"topic_id":  params.TopicID,
			"username":  params.Username,
		},
	})
	return err
}

// GetUserByID gets a user by ID
func (c *Client) GetUserByID(id string) (*UserData, error) {
	resp, err := c.doRequest(Request{
		Query:      "user",
		Condition:  "specific",
		Sort:       "id",
		Parameters: map[string]interface{}{"id": id},
	})
	if err != nil {
		return nil, err
	}

	var data UserData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &data, nil
}

// GetUserByEmail gets a user by email
func (c *Client) GetUserByEmail(email string) (*UserData, error) {
	resp, err := c.doRequest(Request{
		Query:      "user",
		Condition:  "specific",
		Sort:       "email",
		Parameters: map[string]interface{}{"email": email},
	})
	if err != nil {
		return nil, err
	}

	var data UserData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &data, nil
}

// CreateUserParams contains parameters for creating a user
type CreateUserParams struct {
	Name           string
	Email          string
	Password       string
	Phone          string
	Timezone       string
	OrgID          int
	DefaultEmailID int
	Status         int
}

// CreateUser creates a new user
func (c *Client) CreateUser(params CreateUserParams) (int, error) {
	resp, err := c.doRequest(Request{
		Query:     "user",
		Condition: "add",
		Parameters: map[string]interface{}{
			"name":             params.Name,
			"email":            params.Email,
			"password":         params.Password,
			"phone":            params.Phone,
			"timezone":         params.Timezone,
			"org_id":           params.OrgID,
			"default_email_id": params.DefaultEmailID,
			"status":           params.Status,
		},
	})
	if err != nil {
		return 0, err
	}

	var userID int
	if err := json.Unmarshal(resp.Data, &userID); err != nil {
		return 0, fmt.Errorf("failed to parse user ID: %w", err)
	}

	return userID, nil
}

// GetDepartments gets all departments
func (c *Client) GetDepartments() (*DepartmentData, error) {
	resp, err := c.doRequest(Request{
		Query:      "department",
		Condition:  "all",
		Sort:       "all",
		Parameters: map[string]interface{}{},
	})
	if err != nil {
		return nil, err
	}

	var data DepartmentData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse department data: %w", err)
	}

	return &data, nil
}

// GetTopics gets all help topics
func (c *Client) GetTopics() (*TopicData, error) {
	resp, err := c.doRequest(Request{
		Query:      "topics",
		Condition:  "all",
		Sort:       "all",
		Parameters: map[string]interface{}{},
	})
	if err != nil {
		return nil, err
	}

	var data TopicData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse topic data: %w", err)
	}

	return &data, nil
}

// GetSLAs gets all SLA plans
func (c *Client) GetSLAs() (*SLAData, error) {
	resp, err := c.doRequest(Request{
		Query:      "sla",
		Condition:  "all",
		Sort:       "all",
		Parameters: map[string]interface{}{},
	})
	if err != nil {
		return nil, err
	}

	var data SLAData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse SLA data: %w", err)
	}

	return &data, nil
}

// SearchTicketsByEmail searches tickets by user email
func (c *Client) SearchTicketsByEmail(email string) (*TicketData, *User, error) {
	// First get the user
	userData, err := c.GetUserByEmail(email)
	if err != nil {
		return nil, nil, err
	}

	if len(userData.Users) == 0 {
		return &TicketData{Total: 0, Tickets: [][]Ticket{}}, nil, nil
	}

	user := userData.Users[0]

	// Get all tickets
	allTickets, err := c.GetTicketsByStatus(0)
	if err != nil {
		return nil, &user, err
	}

	// Filter by user ID
	var filtered [][]Ticket
	for _, ticketGroup := range allTickets.Tickets {
		for _, t := range ticketGroup {
			if t.UserID == user.UserID {
				filtered = append(filtered, ticketGroup)
				break
			}
		}
	}

	return &TicketData{
		Total:   len(filtered),
		Tickets: filtered,
	}, &user, nil
}
