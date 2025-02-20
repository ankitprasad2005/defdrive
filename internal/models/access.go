package models

type Access struct {
    AccessID string `json:"access_id" pg:",pk"`
    Username string `json:"username"`
    FileID   string `json:"file_id"`
    Link     string `json:"link"`
    Expiry   string `json:"expiry"`
}