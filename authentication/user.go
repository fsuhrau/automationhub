package authentication

// User is the signed-in user of a request.
type User struct {
	// ID is the id of the hub user (models.User.ID). It is the only part of
	// the session that identifies the user; the other fields are labels.
	ID      uint   `json:"id"`
	Login   string `json:"login"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	URL     string `json:"url"`
	Role    string `json:"role"`
}
