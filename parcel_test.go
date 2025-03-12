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
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, p.Client)
	require.Equal(t, parcel.Status, p.Status)
	require.Equal(t, parcel.Address, p.Address)
	require.Equal(t, parcel.CreatedAt, p.CreatedAt)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()      // настройте подключение к БД
	tx, err := db.Begin() // Начало транзакции
	if err != nil {
		fmt.Println("Ошибка начала транзакции:", err)
		return
	}
	defer tx.Rollback() // Откат изменений

	store := NewParcelStore(db) // Инициализация store
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close() // настройте подключение к БД

	tx, err := db.Begin() // Начало транзакции
	if err != nil {
		fmt.Println("Ошибка начала транзакции:", err)
		return
	}
	defer tx.Rollback() // Откат изменений

	store := NewParcelStore(db) // Инициализация store
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tx.Rollback()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка добавления посылки")
		require.NotZero(t, id, "Идентификатор посылки не должен быть нулевым") // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка получения посылок")
	require.Len(t, storedParcels, len(parcels), "Количество посылок не совпадает") // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists, "Посылка с номером %d не найдена", parcel.Number)
		require.Equal(t, expectedParcel.Client, parcel.Client, "Неверный клиент")
		require.Equal(t, expectedParcel.Status, parcel.Status, "Неверный статус")
		require.Equal(t, expectedParcel.Address, parcel.Address, "Неверный адрес")
		require.Equal(t, expectedParcel.CreatedAt, parcel.CreatedAt, "Неверное время создания")
	}
}
