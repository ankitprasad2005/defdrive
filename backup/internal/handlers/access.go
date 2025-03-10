package handlers

import (
    "defdrive/internal/db"
    "defdrive/internal/models"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

func CreateAccess(w http.ResponseWriter, r *http.Request) {
    var access models.Access
    if err := json.NewDecoder(r.Body).Decode(&access); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := createAccess(&access); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func GetAccess(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    accessID := vars["accessID"]
    access, err := getAccessByID(accessID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(access)
}

func createAccess(access *models.Access) error {
    _, err := db.DB.Model(access).Insert()
    return err
}

func getAccessByID(accessID string) (*models.Access, error) {
    access := &models.Access{}
    err := db.DB.Model(access).Where("access_id = ?", accessID).Select()
    return access, err
}