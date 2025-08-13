package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка при добавлении посылки: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении ID: %w", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	err := s.db.QueryRow(
		`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`,
		number,
	).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`,
		client,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе посылок: %w", err)
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(
			&p.Number,
			&p.Client,
			&p.Status,
			&p.Address,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении данных: %w", err)
		}
		parcels = append(parcels, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %w", err)
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	if status == "" {
		return errors.New("статус не может быть пустым")
	}

	result, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?`,
		status,
		number,
	)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("посылка с номером %d не найдена", number)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	if address == "" {
		return errors.New("адрес не может быть пустым")
	}

	var currentStatus string
	err := s.db.QueryRow(
		"SELECT status FROM parcel WHERE number = ?",
		number,
	).Scan(&currentStatus)

	if err != nil {
		return fmt.Errorf("ошибка при проверке статуса посылки: %w", err)
	}

	if currentStatus != "registered" {
		return fmt.Errorf("можно изменить адрес только для посылок со статусом 'registered'")
	}

	result, err := s.db.Exec(
		"UPDATE parcel SET address = ? WHERE number = ?",
		address,
		number,
	)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении адреса: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("посылка с номером %d не найдена", number)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	var currentStatus string
	err := s.db.QueryRow(
		"SELECT status FROM parcel WHERE number = ?",
		number,
	).Scan(&currentStatus)

	if err != nil {
		return fmt.Errorf("ошибка при проверке статуса посылки: %w", err)
	}

	if currentStatus != "registered" {
		return fmt.Errorf("можно удалять только посылки со статусом 'registered'")
	}

	result, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = ?",
		number,
	)
	if err != nil {
		return fmt.Errorf("ошибка при удалении посылки: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("посылка с номером %d не найдена", number)
	}

	return nil
}
