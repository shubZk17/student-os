package config

import "testing"

func TestProductionRejectsDevSecrets(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("DB_PASSWORD", "x")
	t.Setenv("DB_SSLMODE", "require")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ValidateJWT() == nil {
		t.Fatal("expected error for default JWT_SECRET in production")
	}

	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	if cfg, err = Load(); err != nil || cfg.ValidateJWT() != nil {
		t.Fatalf("unexpected error: %v / %v", err, cfg.ValidateJWT())
	}

	cfg.DBPassword = "p@ss/word"
	if got, want := cfg.DatabaseDSN(), "postgres://studentos:p%40ss%2Fword@localhost:5432/studentos_db?sslmode=require"; got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}
}
