package storage

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"testing"
)

func Test_dbStorage_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	ctx := context.Background()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewDBStorage(ctx, db)

	row := mock.NewRows([]string{"original_url", "is_deleted"}).AddRow("original_url", false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT original_url, is_deleted FROM urls WHERE short_url=$1")).
		WithArgs("short_url1").
		WillReturnRows(row)
	want := "original_url"
	got, _ := s.Get(ctx, "short_url1")
	if want != got {
		t.Errorf("got %v want %v", got, want)
	}
}

func Test_dbStorage_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	ctx := context.Background()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewDBStorage(ctx, db)

	row := mock.NewRows([]string{"short_url", "original_url"}).AddRow("short_url", "original_url")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT short_url, original_url FROM urls WHERE user_id=$1")).
		WithArgs("1").
		WillReturnRows(row)
	want := map[string]string{"short_url": "original_url"}
	got, _ := s.GetAll(ctx, "1")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func Test_dbStorage_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	ctx := context.Background()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s := NewDBStorage(ctx, db)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE urls set is_deleted=true WHERE short_url=$1 and user_id=$2`)).
		WithArgs("1", "1").
		WillReturnResult(sqlmock.NewResult(1, 1)).
		WillReturnError(nil)
	err = s.Delete("1", "1")
	assert.NoError(t, err)
}

func Test_dbStorage_Add(t *testing.T) {
	db, mock, err := sqlmock.New()
	ctx := context.Background()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s := NewDBStorage(ctx, db)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls(short_url, original_url, user_id) VALUES ($1, $2, $3)`)).
		WithArgs("1", "1", "1").
		WillReturnResult(sqlmock.NewResult(1, 1)).
		WillReturnError(nil)
	err = s.Add(ctx, "1", "1", "1")
	assert.NoError(t, err)
}

func Test_dbStorage_AddBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	ctx := context.Background()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s := NewDBStorage(ctx, db)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls(short_url, original_url, user_id) VALUES($1, $2, $3) ON CONFLICT DO NOTHING`)).
		WithArgs("1", "1", "1").
		WillReturnResult(sqlmock.NewResult(1, 1)).
		WillReturnError(nil)
	mock.ExpectCommit()
	err = s.AddBatch(ctx, map[string]string{"1": "1"}, "1")
	assert.NoError(t, err)
}
