package storage

import (
	"context"
	"errors"
)

type MocksqlStorage struct {
	migrations []IMigration
}

func NewMocksqlStorage() *MocksqlStorage {
	return &MocksqlStorage{
		migrations: []IMigration{},
	}
}

func (m *MocksqlStorage) Connect(_ context.Context) error {
	return nil
}

func (m *MocksqlStorage) Close() error {
	return nil
}

func (m *MocksqlStorage) Lock(_ context.Context) error {
	return nil
}

func (m *MocksqlStorage) Unlock(_ context.Context) error {
	return nil
}

func (m *MocksqlStorage) InsertMigration(ctx context.Context, migration IMigration) error {
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

func (m *MocksqlStorage) UpdateMigration(ctx context.Context, migration IMigration) error {
	for _, m := range m.migrations {
		if m.GetVersion() == migration.GetVersion() && m.GetName() == migration.GetName() {
			m.SetStatus(migration.GetStatus())
			m.SetStatusChangeTime(migration.GetStatusChangeTime())
			return nil
		}
	}
	return errors.New("migration not found")
}

func (m *MocksqlStorage) Migrate(ctx context.Context, sql string) error {
	return nil
}

func (m *MocksqlStorage) SelectMigrations(ctx context.Context) ([]IMigration, error) {
	if len(m.migrations) == 0 {
		return nil, errors.New("no migrations found")
	}
	return m.migrations, nil
}

func (m *MocksqlStorage) SelectLastMigrationByStatus(ctx context.Context, status string) (IMigration, error) {
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if m.migrations[i].GetStatus() == status {
			return m.migrations[i], nil
		}
	}
	return nil, errors.New("no migrations found with status " + status)
}

func (m *MocksqlStorage) DeleteMigrations(ctx context.Context) error {
	m.migrations = []IMigration{}
	return nil
}
