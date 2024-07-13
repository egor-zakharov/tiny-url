package storage

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
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
