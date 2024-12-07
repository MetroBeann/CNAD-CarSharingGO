// Path: services/vehicle-service/repository/reservation_repository.go
package repository

import (
    "database/sql"
    "errors"
    "time"

    "vehicle-service/models"
)

type ReservationRepository struct {
    DB *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
    return &ReservationRepository{DB: db}
}

func (r *ReservationRepository) CreateReservation(reservation *models.Reservation) error {
    query := `
        INSERT INTO reservations (user_id, vehicle_id, start_time, end_time, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $6)
        RETURNING id
    `
    
    err := r.DB.QueryRow(
        query,
        reservation.UserID,
        reservation.VehicleID,
        reservation.StartTime,
        reservation.EndTime,
        "Active",
        time.Now(),
    ).Scan(&reservation.ID)

    if err != nil {
        return err
    }

    // Update vehicle status to In-Use
    updateQuery := `UPDATE vehicles SET status = 'In-Use' WHERE id = $1`
    _, err = r.DB.Exec(updateQuery, reservation.VehicleID)
    
    return err
}

func (r *ReservationRepository) GetUserReservations(userID int) ([]models.Reservation, error) {
    query := `
        SELECT r.id, r.user_id, r.vehicle_id, r.start_time, r.end_time, r.status,
               r.created_at, r.updated_at,
               v.model, v.type, v.location
        FROM reservations r
        JOIN vehicles v ON r.vehicle_id = v.id
        WHERE r.user_id = $1
        ORDER BY r.start_time DESC
    `

    rows, err := r.DB.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var reservations []models.Reservation
    for rows.Next() {
        var r models.Reservation
        r.Vehicle = &models.Vehicle{}
        err := rows.Scan(
            &r.ID, &r.UserID, &r.VehicleID, &r.StartTime, &r.EndTime, &r.Status,
            &r.CreatedAt, &r.UpdatedAt,
            &r.Vehicle.Model, &r.Vehicle.Type, &r.Vehicle.Location,
        )
        if err != nil {
            return nil, err
        }
        reservations = append(reservations, r)
    }

    return reservations, nil
}

func (r *ReservationRepository) UpdateReservation(id int, userID int, updates models.UpdateReservationRequest) error {
    tx, err := r.DB.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Check if reservation exists and belongs to user
    var status string
    err = tx.QueryRow(
        "SELECT status FROM reservations WHERE id = $1 AND user_id = $2",
        id, userID,
    ).Scan(&status)
    if err == sql.ErrNoRows {
        return errors.New("reservation not found or unauthorized")
    }
    if err != nil {
        return err
    }

    if status != "Active" {
        return errors.New("cannot modify non-active reservation")
    }

    // Update reservation
    query := `
        UPDATE reservations 
        SET start_time = COALESCE($1, start_time),
            end_time = COALESCE($2, end_time),
            updated_at = $3
        WHERE id = $4 AND user_id = $5
    `

    result, err := tx.Exec(
        query,
        updates.StartTime,
        updates.EndTime,
        time.Now(),
        id,
        userID,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return errors.New("reservation not found")
    }

    return tx.Commit()
}

func (r *ReservationRepository) CancelReservation(id int, userID int) error {
    tx, err := r.DB.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Get vehicle ID and check if reservation is active
    var vehicleID int
    var status string
    err = tx.QueryRow(
        "SELECT vehicle_id, status FROM reservations WHERE id = $1 AND user_id = $2",
        id, userID,
    ).Scan(&vehicleID, &status)
    if err == sql.ErrNoRows {
        return errors.New("reservation not found or unauthorized")
    }
    if err != nil {
        return err
    }

    if status != "Active" {
        return errors.New("reservation is already cancelled or completed")
    }

    // Update reservation status
    _, err = tx.Exec(
        "UPDATE reservations SET status = 'Cancelled', updated_at = $1 WHERE id = $2",
        time.Now(), id,
    )
    if err != nil {
        return err
    }

    // Update vehicle status back to Available
    _, err = tx.Exec(
        "UPDATE vehicles SET status = 'Available' WHERE id = $1",
        vehicleID,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}