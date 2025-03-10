package models

type File struct {
    FileID   string `json:"file_id" pg:",pk"`
    Location string `json:"location"`
    Hash     string `json:"hash"`
    Name     string `json:"name"`
    Size     int64  `json:"size"`
}