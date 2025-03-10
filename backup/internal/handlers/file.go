package handlers

import (
    "defdrive/internal/db"
    "defdrive/internal/models"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

func CreateFile(w http.ResponseWriter, r *http.Request) {
    var file models.File
    if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := createFile(&file); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileID := vars["fileID"]
    file, err := getFileByID(fileID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(file)
}

func createFile(file *models.File) error {
    _, err := db.DB.Model(file).Insert()
    return err
}

func getFileByID(fileID string) (*models.File, error) {
    file := &models.File{}
    err := db.DB.Model(file).Where("file_id = ?", fileID).Select()
    return file, err
}