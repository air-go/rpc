package gorm

import (
	"testing"

	"gorm.io/gorm"
)

func TestNewMemoryDB(t *testing.T) {
	tests := []struct {
		name string
		want *gorm.DB
	}{
		{
			name: "success",
			want: &gorm.DB{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryDB(); got == nil {
				t.Errorf("NewMemoryDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloseMemoryDB(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	args1 := NewMemoryDB()
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{db: args1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CloseMemoryDB(tt.args.db)
		})
	}
}
