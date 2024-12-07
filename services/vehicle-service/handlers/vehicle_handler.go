// Path: services/vehicle-service/handlers/vehicle_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "vehicle-service/models"
    "vehicle-service/repository"
)

type VehicleHandler struct {
    VehicleRepo     *repository.VehicleRepository
    ReservationRepo *repository.ReservationRepository
}

func NewVehicleHandler(vRepo *repository.VehicleRepository, rRepo *repository.ReservationRepository) *VehicleHandler {
    return &VehicleHandler{
        VehicleRepo:     vRepo,
        ReservationRepo: rRepo,
    }
}

func (h *VehicleHandler) GetAvailableVehicles(w http.ResponseWriter, r *http.Request) {
    var req models.AvailabilityRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if req.StartTime.IsZero() || req.EndTime.IsZero() {
        http.Error(w, "Start time and end time are required", http.StatusBadRequest)
        return
    }

    if req.EndTime.Before(req.StartTime) {
        http.Error(w, "End time must be after start time", http.StatusBadRequest)
        return
    }

    vehicles, err := h.VehicleRepo.GetAvailableVehicles(req.StartTime, req.EndTime)
    if err != nil {
        http.Error(w, "Failed to get available vehicles", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(vehicles)
}

func (h *VehicleHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
    var req models.ReservationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Get user ID from context (set by auth middleware)
    userID := r.Context().Value("user_id").(int)

    reservation := &models.Reservation{
        UserID:    userID,
        VehicleID: req.VehicleID,
        StartTime: req.StartTime,
        EndTime:   req.EndTime,
    }

    if err := h.ReservationRepo.CreateReservation(reservation); err != nil {
        http.Error(w, "Failed to create reservation: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(reservation)
}

func (h *VehicleHandler) GetUserReservations(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(int)

    reservations, err := h.ReservationRepo.GetUserReservations(userID)
    if err != nil {
        http.Error(w, "Failed to get reservations", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(reservations)
}

func (h *VehicleHandler) UpdateReservation(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    reservationID, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid reservation ID", http.StatusBadRequest)
        return
    }

    userID := r.Context().Value("user_id").(int)

    var req models.UpdateReservationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.ReservationRepo.UpdateReservation(reservationID, userID, req); err != nil {
        http.Error(w, "Failed to update reservation: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Reservation updated successfully",
    })
}

func (h *VehicleHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    reservationID, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid reservation ID", http.StatusBadRequest)
        return
    }

    userID := r.Context().Value("user_id").(int)

    if err := h.ReservationRepo.CancelReservation(reservationID, userID); err != nil {
        http.Error(w, "Failed to cancel reservation: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Reservation cancelled successfully",
    })
}