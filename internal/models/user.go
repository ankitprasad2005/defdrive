package models

type User struct {
    Username string `json:"username" pg:",pk"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}