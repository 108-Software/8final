package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
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
    	require.True(t, exists, "Посылка с номером %d не найдена в исходных данных", p.Number)
		
    	require.Equal(t, original, p, "Посылка с номером %d не соответствует ожидаемой", p.Number,
    	    "При сравнении игнорируется поле Number, так как оно может различаться")
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
    require.NoError(t, err, "Не удалось подключиться к базе данных")
    defer db.Close()

    store := NewParcelStore(db)

    parcel := getTestParcel()
    id, err := store.Add(parcel)
    require.NoError(t, err)
    require.NotZero(t, id)

    newAddress := "new test address"
    err = store.SetAddress(id, newAddress)
    require.NoError(t, err)

    updatedParcel, err := store.Get(id)
    require.NoError(t, err)

    expectedParcel := parcel
    expectedParcel.Number = id 
    expectedParcel.Address = newAddress

    require.Equal(t, expectedParcel, updatedParcel, "Посылка после обновления адреса не соответствует ожидаемой")
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Не удалось подключиться к базе данных")
	defer db.Close()

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

	expectedParcel := parcel          
	expectedParcel.Number = updatedParcel.Number
	expectedParcel.Status = newStatus

	require.Equal(t, expectedParcel, updatedParcel, "Посылка после обновления статуса не соответствует ожидаемой")
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Не удалось подключиться к базе данных")
	defer db.Close()

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
   		require.NoError(t, err, "Не удалось добавить посылку")
   		require.NotZero(t, id, "Ожидался ненулевой ID посылки")

   		parcels[i].Number = id
   		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Не удалось получить посылки клиента")
	require.Equal(t, len(parcels), len(storedParcels), "Количество полученных посылок не соответствует ожидаемому")

	for _, parcel := range storedParcels {
    	original, exists := parcelMap[parcel.Number]
    	require.True(t, exists, "Посылка с номером %d не найдена в исходных данных", parcel.Number)
		
    	assert.Equal(t, original.Client, parcel.Client, "Несоответствие Client для посылки %d", parcel.Number)
    	assert.Equal(t, original.Status, parcel.Status, "Несоответствие Status для посылки %d", parcel.Number)
    	assert.Equal(t, original.Address, parcel.Address, "Несоответствие Address для посылки %d", parcel.Number)
    	assert.Equal(t, original.CreatedAt, parcel.CreatedAt, "Несоответствие CreatedAt для посылки %d", parcel.Number)
	}
}
