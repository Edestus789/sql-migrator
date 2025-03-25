package storage

import (
	"context"
	"errors"
)

type MockSQLStorage struct {
	migrations []IMigration
}

func NewMockSQLStorage() *MockSQLStorage {
	return &MockSQLStorage{
		migrations: []IMigration{},
	}
}

func (m *MockSQLStorage) Connect(_ context.Context) error {
	return nil
}

func (m *MockSQLStorage) Close() error {
	return nil
}

func (m *MockSQLStorage) Lock(_ context.Context) error {
	return nil
}

func (m *MockSQLStorage) Unlock(_ context.Context) error {
	return nil
}

func (m *MockSQLStorage) InsertMigration(_ context.Context, migration IMigration) error {
	for _, m := range m.migrations {
		if m.GetVersion() == migration.GetVersion() && m.GetName() == migration.GetName() {
			m.SetStatus(migration.GetStatus())
			m.SetStatusChangeTime(migration.GetStatusChangeTime())
			m.SetVersion(migration.GetVersion())
			m.SetName(migration.GetName())
			return nil
		}
	}
	m.migrations = append(m.migrations, migration)
	return nil
}

func (m *MockSQLStorage) UpdateMigration(_ context.Context, migration IMigration) error {
	for _, m := range m.migrations {
		if m.GetVersion() == migration.GetVersion() && m.GetName() == migration.GetName() {
			m.SetStatus(migration.GetStatus())
			m.SetStatusChangeTime(migration.GetStatusChangeTime())
			return nil
		}
	}
	return errors.New("migration not found")
}

func (m *MockSQLStorage) Migrate(_ context.Context, sql string) error {
	return nil
}

func (m *MockSQLStorage) SelectMigrations(_ context.Context) ([]IMigration, error) {
	if len(m.migrations) == 0 {
		return nil, errors.New("no migrations found")
	}
	return m.migrations, nil
}

func (m *MockSQLStorage) SelectLastMigrationByStatus(_ context.Context, status string) (IMigration, error) {
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if m.migrations[i].GetStatus() == status {
			return m.migrations[i], nil
		}
	}
	return nil, errors.New("no migrations found with status " + status)
}

func (m *MockSQLStorage) DeleteMigrations(_ context.Context) error {
	m.migrations = []IMigration{}
	return nil
}
