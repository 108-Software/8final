package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	clientID := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = clientID
	}

	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id, "ID посылки не должен быть 0")

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels), "Количество посылок не совпадает")

	for _, p := range storedParcels {
		original, exists := parcelMap[p.Number]
		require.True(t, exists, "Посылка не найдена в исходных данных")
		require.Equal(t, original.Client, p.Client)
		require.Equal(t, original.Status, p.Status)
		require.Equal(t, original.Address, p.Address)
		require.Equal(t, original.CreatedAt, p.CreatedAt)
	}

	for id := range parcelMap {
		err := store.Delete(id)
		require.NoError(t, err)

		_, err = store.Get(id)
		require.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	store := NewParcelStore(db)
	defer db.Close()

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	updatedParcel, err := store.Get(id)
	require.NoError(t, err)

	require.Equal(t, newAddress, updatedParcel.Address)
	require.Equal(t, parcel.Client, updatedParcel.Client)
	require.Equal(t, parcel.Status, updatedParcel.Status)
	require.Equal(t, parcel.CreatedAt, updatedParcel.CreatedAt)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	store := NewParcelStore(db)
	defer db.Close()

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err, "Failed to add test parcel")
	require.NotZero(t, id, "Expected non-zero parcel ID")

	initialParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusRegistered, initialParcel.Status)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err, "Failed to update status")

	updatedParcel, err := store.Get(id)
	require.NoError(t, err)

	require.Equal(t, newStatus, updatedParcel.Status, "Status was not updated")
	require.Equal(t, parcel.Client, updatedParcel.Client, "Client was changed unexpectedly")
	require.Equal(t, parcel.Address, updatedParcel.Address, "Address was changed unexpectedly")
	require.Equal(t, parcel.CreatedAt, updatedParcel.CreatedAt, "CreatedAt was changed unexpectedly")
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatalf("Failed to add parcel: %v", err)
		}
		if id == 0 {
			t.Fatal("Expected non-zero ID")
		}

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatalf("Failed to get parcels by client: %v", err)
	}
	if len(storedParcels) != len(parcels) {
		t.Fatalf("Expected %d parcels, got %d", len(parcels), len(storedParcels))
	}

	for _, parcel := range storedParcels {
		original, exists := parcelMap[parcel.Number]
		if !exists {
			t.Fatalf("Parcel with number %d not found in original data", parcel.Number)
		}

		if parcel.Client != original.Client {
			t.Errorf("Client mismatch: expected %d, got %d", original.Client, parcel.Client)
		}
		if parcel.Status != original.Status {
			t.Errorf("Status mismatch: expected %s, got %s", original.Status, parcel.Status)
		}
		if parcel.Address != original.Address {
			t.Errorf("Address mismatch: expected %s, got %s", original.Address, parcel.Address)
		}
		if parcel.CreatedAt != original.CreatedAt {
			t.Errorf("CreatedAt mismatch: expected %s, got %s", original.CreatedAt, parcel.CreatedAt)
		}
	}
}
