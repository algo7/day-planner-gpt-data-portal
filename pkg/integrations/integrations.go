package integrations

// Email is a struct to hold the email data
type Email struct {
	Subject          string `json:"subject"`
	Body             string `json:"body"`
	Sender           string `json:"sender"`
	RecievedDateTime string `json:"recievedDateTime"`
}
