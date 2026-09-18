package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func resetTemplateCacheForTest() {
	templateCache = cache.NewMapCache(func(t Template) int { return t.ID })
	templateCache.AddIndex("name", func(t Template) string { return t.Name })
}

func setupTemplateTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "template_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Template{}); err != nil {
		t.Fatalf("auto migrate templates: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetTemplateCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetTemplateCacheForTest()
		testutil.CloseDB(t, db)
	})
}

func TestInferTemplateCategory(t *testing.T) {
	tests := map[string]string{
		"clash.yaml": "clash",
		"surge.conf": "surge",
		"SURGE.CONF": "surge",
		"loon.lcf":   "loon",
		"LOON.LCF":   "loon",
		"rules.txt":  "clash",
	}

	for fileName, want := range tests {
		if got := InferTemplateCategory(fileName); got != want {
			t.Fatalf("InferTemplateCategory(%q)=%q, want %q", fileName, got, want)
		}
	}
}

func TestMigrateTemplatesFromFilesCreatesExpectedCategories(t *testing.T) {
	setupTemplateTestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "clash.yaml"), []byte("proxies: []\n"), 0600); err != nil {
		t.Fatalf("write clash template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "surge.conf"), []byte("[General]\n"), 0600); err != nil {
		t.Fatalf("write surge template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "loon.lcf"), []byte("[General]\n[Proxy]\n[Proxy Group]\n"), 0600); err != nil {
		t.Fatalf("write Loon template: %v", err)
	}

	if err := MigrateTemplatesFromFiles(templateDir); err != nil {
		t.Fatalf("migrate templates: %v", err)
	}

	var templates []Template
	if err := database.DB.Order("name asc").Find(&templates).Error; err != nil {
		t.Fatalf("query templates: %v", err)
	}
	if len(templates) != 3 {
		t.Fatalf("expected 3 templates, got %d", len(templates))
	}

	got := map[string]string{}
	for _, tmpl := range templates {
		got[tmpl.Name] = tmpl.Category
	}
	if got["clash.yaml"] != "clash" {
		t.Fatalf("expected clash.yaml category clash, got %q", got["clash.yaml"])
	}
	if got["surge.conf"] != "surge" {
		t.Fatalf("expected surge.conf category surge, got %q", got["surge.conf"])
	}
	if got["loon.lcf"] != "loon" {
		t.Fatalf("expected loon.lcf category loon, got %q", got["loon.lcf"])
	}
}

func TestMigrateTemplatesFromFilesRepairsInvalidCategory(t *testing.T) {
	setupTemplateTestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "surge.conf"), []byte("[General]\n"), 0600); err != nil {
		t.Fatalf("write surge template: %v", err)
	}

	existing := Template{
		Name:     "surge.conf",
		Category: "unknown",
	}
	if err := database.DB.Create(&existing).Error; err != nil {
		t.Fatalf("create existing template: %v", err)
	}

	if err := MigrateTemplatesFromFiles(templateDir); err != nil {
		t.Fatalf("migrate templates: %v", err)
	}

	var repaired Template
	if err := database.DB.Where("name = ?", "surge.conf").First(&repaired).Error; err != nil {
		t.Fatalf("query repaired template: %v", err)
	}
	if repaired.Category != "surge" {
		t.Fatalf("expected repaired category surge, got %q", repaired.Category)
	}
}

func TestMigrateTemplatesFromFilesRepairsLegacyLoonCategory(t *testing.T) {
	setupTemplateTestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "legacy.lcf"), []byte("[General]\n[Proxy]\n[Proxy Group]\n"), 0o600); err != nil {
		t.Fatalf("write legacy Loon template: %v", err)
	}
	existing := Template{Name: "legacy.lcf", Category: TemplateCategoryClash}
	if err := database.DB.Create(&existing).Error; err != nil {
		t.Fatalf("create legacy template metadata: %v", err)
	}

	if err := MigrateTemplatesFromFiles(templateDir); err != nil {
		t.Fatalf("migrate templates: %v", err)
	}
	var repaired Template
	if err := database.DB.Where("name = ?", "legacy.lcf").First(&repaired).Error; err != nil {
		t.Fatalf("query repaired Loon template: %v", err)
	}
	if repaired.Category != TemplateCategoryLoon {
		t.Fatalf("category = %q, want %q", repaired.Category, TemplateCategoryLoon)
	}
}

func TestRepairLegacyLoonTemplateAssignments(t *testing.T) {
	db := testutil.OpenMemoryDB(t, "legacy_loon_assignments")
	t.Cleanup(func() { testutil.CloseDB(t, db) })
	if err := db.AutoMigrate(&Template{}, &Subcription{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Create(&Template{Name: "forLoon.lcf", Category: TemplateCategoryClash}).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	subscription := Subcription{Name: "legacy-loon", Config: `{"clash":"./template/forLoon.lcf","surge":"./template/surge.conf","cert":true}`}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	if err := repairLegacyLoonTemplateAssignments(db); err != nil {
		t.Fatalf("repair legacy assignments: %v", err)
	}
	var template Template
	if err := db.Where("name = ?", "forLoon.lcf").First(&template).Error; err != nil {
		t.Fatalf("query template: %v", err)
	}
	if template.Category != TemplateCategoryLoon {
		t.Fatalf("template category = %q, want loon", template.Category)
	}
	var repaired Subcription
	if err := db.First(&repaired, subscription.ID).Error; err != nil {
		t.Fatalf("query subscription: %v", err)
	}
	if !strings.Contains(repaired.Config, `"loon":"./template/forLoon.lcf"`) {
		t.Fatalf("repaired config missing Loon assignment: %s", repaired.Config)
	}
	if !strings.Contains(repaired.Config, `"clash":"./template/clash.yaml"`) {
		t.Fatalf("repaired config missing restored Clash template: %s", repaired.Config)
	}
	if !strings.Contains(repaired.Config, `"cert":true`) {
		t.Fatalf("repaired config lost unrelated values: %s", repaired.Config)
	}
}
