package authentication

type User struct {
	Login   string `json:"login"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	URL     string `json:"url"`
	Role    string `json:"role"`
}
