package domain

type Venue struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	City     string `json:"city"`
	Category string `json:"category"` // cafe | restaurant | park | cultural
	Address  string `json:"address"`
	Phone    string `json:"phone,omitempty"`
}
