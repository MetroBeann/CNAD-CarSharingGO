// Path: services/vehicle-service/models/models.go
package models

import (
    "time"
)

type Vehicle struct {
    ID          int       `json:"id"`
    Model       string    `json:"model"`
    Type        string    `json:"type"`
    Status      string    `json:"status"` // Available, In-Use, Maintenance
    Location    string    `json:"location"`
    ChargeLevel int       `json:"charge_level,omitempty"`
    Cleanliness string    `json:"cleanliness,omitempty"` // Clean, Needs-Cleaning
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Reservation struct {
    ID         int       `json:"id"`
    UserID     int       `json:"user_id"`
    VehicleID  int       `json:"vehicle_id"`
    StartTime  time.Time `json:"start_time"`
    EndTime    time.Time `json:"end_time"`
    Status     string    `json:"status"` // Active, Completed, Cancelled
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    Vehicle    *Vehicle  `json:"vehicle,omitempty"`
}

type AvailabilityRequest struct {
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
}

type ReservationRequest struct {
    VehicleID int       `json:"vehicle_id"`
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
}

type UpdateReservationRequest struct {
    StartTime *time.Time `json:"start_time,omitempty"`
    EndTime   *time.Time `json:"end_time,omitempty"`
}

type VehicleStatusUpdate struct {
    Status      *string `json:"status,omitempty"`
    ChargeLevel *int    `json:"charge_level,omitempty"`
    Cleanliness *string `json:"cleanliness,omitempty"`
    Location    *string `json:"location,omitempty"`
}