package querycache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/google/uuid"
)

func TestInMemoryAdapterCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)

	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	adapter, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	expectedStore := map[string]*querycache.Adapter{}
	expectedStore[id] = expected

	expect.Equal(t, expectedStore, store.Store)
}

func TestInMemoryAdapterDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)

	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	store.Store[id] = expected

	adapter, err := store.Delete(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	expectedStore := map[string]*querycache.Adapter{}
	expect.Equal(t, expectedStore, store.Store)
}

func TestInMemoryAdapterGet(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	adapter, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)
}

func TestInMemoryAdapterUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test snowdapter",
		Type:      "snowflake",
		Options:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	newName := "test snowdapter"
	newType := "snowflake"
	newOptions := ""
	adapter, err := store.Update(id, &querycache.UpdateAdapter{
		Name:    &newName,
		Type:    &newType,
		Options: &newOptions,
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	adapter, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)
}

func TestInMemoryAdapterList(t *testing.T) {
	t.Parallel()

	store := querycache.NewInMemoryAdapterStore(
		&querycache.RealClock{}, &querycache.UUIDGenerator{})

	expectedAdapters := []*querycache.Adapter{}

	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("name %v;", i)
		q := &querycache.CreateAdapter{Name: s}
		adapter, err := store.Create(q)
		expect.Ok(t, err)

		expectedAdapters = append(expectedAdapters, adapter)
	}

	_, err := store.List(0, 1)
	expect.Error(t, err)

	adapters, err := store.List(1, 10)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters, adapters)

	adapters, err = store.List(2, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters[3:6], adapters)

	adapters, err = store.List(4, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters[9:10], adapters)

	adapters, err = store.List(10, 3)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Adapter{}, adapters)

	adapters, err = store.List(1, 30)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters, adapters)
}
