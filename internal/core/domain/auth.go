package domain

type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID    int64
	Login string
	Hash  string
}
