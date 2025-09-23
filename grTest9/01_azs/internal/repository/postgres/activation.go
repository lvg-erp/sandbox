package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"fuelazs/internal/usecase/models"
	"strings"
	"time"
)

type Activation struct {
	conn *sql.DB
}

func NewActivation(conn *sql.DB) *Activation {
	return &Activation{conn: conn}
}

// GetActivation метод получения данных из таблицы Activation (одна запись)
func (a *Activation) GetActivation() (*models.ActivationData, error) {
	var data models.ActivationData

	const querySQL = `
	SELECT
		last_modification,
		kazs_api_key,
		reset_password,
		kazs_id,
		url,
		config_hash,
		kazs_number,
		kazs_timezone,
		ntp_server,
		support_number,
		logo
	FROM Activation
	LIMIT 1;
	`

	row := a.conn.QueryRow(querySQL)
	err := row.Scan(
		&data.LastModification,
		&data.KazsAPIKey,
		&data.ResetPassword,
		&data.KazsID,
		&data.URL,
		&data.ConfigHash,
		&data.KazsNumber,
		&data.KazsTimezone,
		&data.NTPServer,
		&data.SupportNumber,
		&data.Logo,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// InsertActivation — добавление новой записи Activation
func (a *Activation) InsertActivation(data models.ActivationData) error {
	const query = `
	INSERT INTO Activation (
		last_modification,
		kazs_api_key,
		reset_password,
		kazs_id,
		url,
		config_hash,
		kazs_number,
		kazs_timezone,
		ntp_server,
		support_number,
		logo
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := a.conn.Exec(query,
		data.LastModification,
		data.KazsAPIKey,
		data.ResetPassword,
		data.KazsID,
		data.URL,
		data.ConfigHash,
		data.KazsNumber,
		data.KazsTimezone,
		data.NTPServer,
		data.SupportNumber,
		data.Logo,
	)
	if err != nil {
		return fmt.Errorf("activation insert error: %w", err)
	}

	return nil
}

// UpdateActivation — обновление существующей записи Activation по казс_id
func (a *Activation) UpdateActivation(data models.UpdateActivationData) error {
	if data.KazsID == "" {
		return errors.New("KazsID cannot be empty")
	}

	updates := []string{}
	args := []interface{}{}
	argPos := 1 // для параметров $1, $2, ...

	if data.ConfigHash != nil {
		updates = append(updates, fmt.Sprintf("config_hash = $%d", argPos))
		args = append(args, *data.ConfigHash)
		argPos++
	}

	if data.KazsNumber != nil {
		updates = append(updates, fmt.Sprintf("kazs_number = $%d", argPos))
		args = append(args, *data.KazsNumber)
		argPos++
	}

	if data.KazsTimezone != nil {
		updates = append(updates, fmt.Sprintf("kazs_timezone = $%d", argPos))
		args = append(args, *data.KazsTimezone)
		argPos++
	}

	if data.NtpServer != nil {
		updates = append(updates, fmt.Sprintf("ntp_server = $%d", argPos))
		args = append(args, *data.NtpServer)
		argPos++
	}

	if data.SupportNumber != nil {
		updates = append(updates, fmt.Sprintf("support_number = $%d", argPos))
		args = append(args, *data.SupportNumber)
		argPos++
	}

	if data.Logo != nil {
		updates = append(updates, fmt.Sprintf("logo = $%d", argPos))
		args = append(args, *data.Logo)
		argPos++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for Activation record")
	}

	// last_modification обновляется на текущее время
	updates = append(updates, fmt.Sprintf("last_modification = $%d", argPos))
	args = append(args, time.Now())
	argPos++

	// WHERE kazs_id = $argPos
	query := fmt.Sprintf("UPDATE Activation SET %s WHERE kazs_id = $%d", strings.Join(updates, ", "), argPos)
	args = append(args, data.KazsID)

	res, err := a.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update Activation record for KazsID %s: %w", data.KazsID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("no Activation record found with KazsID %s", data.KazsID)
	}

	return nil
}
