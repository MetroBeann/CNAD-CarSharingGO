// Path: services/vehicle-service/repository/vehicle_repository.go
package repository

import (
    "database/sql"
    "time"

     "vehicle-service/models"
)

type VehicleRepository struct {
    DB *sql.DB
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
    return &VehicleRepository{DB: db}
}

func (r *VehicleRepository) GetAvailableVehicles(startTime, endTime time.Time) ([]models.Vehicle, error) {
    query := `
        SELECT v.id, v.model, v.type, v.status, v.location, v.charge_level, v.cleanliness,
               v.created_at, v.updated_at
        FROM vehicles v
        WHERE v.status = 'Available'
        AND v.id NOT IN (
            SELECT vehicle_id FROM reservations
            WHERE status = 'Active'
            AND (
                (start_time <= $1 AND end_time >= $1)
                OR (start_time <= $2 AND end_time >= $2)
                OR (start_time >= $1 AND end_time <= $2)
            )
        )
    `
    
    rows, err := r.DB.Query(query, startTime, endTime)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var vehicles []models.Vehicle
    for rows.Next() {
        var v models.Vehicle
        err := rows.Scan(
            &v.ID, &v.Model, &v.Type, &v.Status, &v.Location, 
            &v.ChargeLevel, &v.Cleanliness, &v.CreatedAt, &v.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        vehicles = append(vehicles, v)
    }

    return vehicles, nil
}

func (r *VehicleRepository) GetVehicleByID(id int) (*models.Vehicle, error) {
    query := `
        SELECT id, model, type, status, location, charge_level, cleanliness,
               created_at, updated_at
        FROM vehicles WHERE id = $1
    `
    
    var vehicle models.Vehicle
    err := r.DB.QueryRow(query, id).Scan(
        &vehicle.ID, &vehicle.Model, &vehicle.Type, &vehicle.Status,
        &vehicle.Location, &vehicle.ChargeLevel, &vehicle.Cleanliness,
        &vehicle.CreatedAt, &vehicle.UpdatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &vehicle, nil
}

func (r *VehicleRepository) UpdateVehicleStatus(vehicleID int, status string) error {
    query := `
        UPDATE vehicles 
        SET status = $1, updated_at = $2
        WHERE id = $3
    `
    
    _, err := r.DB.Exec(query, status, time.Now(), vehicleID)
    return err
}