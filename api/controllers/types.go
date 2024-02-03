package controllers

// APIResponse represents the structure of a typical REST API response.
type APIResponse struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}
