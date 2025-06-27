package pkg

import "testing"

func TestParseDSNValid(t *testing.T) {
	dsn := "file:test.db?mode=rwc&cache=shared&_journal_mode=WAL&_busy_timeout=3000&_page_size=4096&_synchronous=NORMAL&_foreign_keys=true"
	cfg, err := ParseDSN(dsn)
	if err != nil {
		t.Fatalf("expected valid DSN, got error: %v", err)
	}
	if cfg.Path != "test.db" {
		t.Errorf("unexpected path: %s", cfg.Path)
	}
	if cfg.Mode != "rwc" {
		t.Errorf("mode parse failed, got %s", cfg.Mode)
	}
	if cfg.Cache != "shared" {
		t.Errorf("cache parse failed, got %s", cfg.Cache)
	}
	if cfg.JournalMode != "WAL" {
		t.Errorf("journal_mode parse failed, got %s", cfg.JournalMode)
	}
	if cfg.BusyTimeout.Milliseconds() != 3000 {
		t.Errorf("busy_timeout parse failed, got %d", cfg.BusyTimeout.Milliseconds())
	}
	if cfg.PageSize != 4096 {
		t.Errorf("page_size parse failed, got %d", cfg.PageSize)
	}
	if cfg.Synchronous != "NORMAL" {
		t.Errorf("synchronous parse failed, got %s", cfg.Synchronous)
	}
	if !cfg.ForeignKeys {
		t.Error("foreign_keys parse failed, expected true")
	}
}

func TestParseDSNInvalid(t *testing.T) {
	invalidDSNs := []string{
		"http://test.db",                   // unsupported scheme
		"file:test.db?mode=invalid",        // invalid mode
		"file:test.db?_page_size=123",      // invalid page size (not power of 2)
		"file:test.db?_busy_timeout=abc",   // invalid busy timeout
	}

	for _, dsn := range invalidDSNs {
		if _, err := ParseDSN(dsn); err == nil {
			t.Errorf("expected error for DSN %q, got nil", dsn)
		}
	}
}

