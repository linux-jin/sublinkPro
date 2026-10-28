package services

import (
	"sublink/internal/testutil"
	"sublink/models"
	"testing"
)

func TestDatabaseMigrationPreservesSmartGroupConditionsAndReadsLegacyColumns(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		name := "with-conditions"
		if legacy {
			name = "legacy-columns"
		}
		t.Run(name, func(t *testing.T) {
			source, target := testutil.OpenMemoryDB(t, "conditions_source"), testutil.OpenMemoryDB(t, "conditions_target")
			t.Cleanup(func() { testutil.CloseDB(t, source); testutil.CloseDB(t, target) })
			if err := source.AutoMigrate(&models.NodeCheckProfile{}, &models.SmartGroup{}); err != nil {
				t.Fatal(err)
			}
			if err := target.AutoMigrate(&models.NodeCheckProfile{}, &models.SmartGroup{}, &models.NodeCheckJob{}); err != nil {
				t.Fatal(err)
			}
			group := models.SmartGroup{ID: 9, Name: "restored", Countries: "GB", CandidateConditions: `{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":"ss"}]}`, MemberConditions: `{"logic":"and","conditions":[{"field":"quality_status","operator":"equals","value":"success"}]}`}
			if err := source.Create(&group).Error; err != nil {
				t.Fatal(err)
			}
			if legacy {
				for _, column := range []string{"candidate_conditions", "member_conditions"} {
					if err := source.Migrator().DropColumn(&models.SmartGroup{}, column); err != nil {
						t.Fatal(err)
					}
				}
			}
			state := &databaseMigrationState{source: source, tx: target, sourceHasProfiles: true, result: &DatabaseMigrationResult{Imported: make(map[string]int)}}
			if err := importNodeCheckProfiles(state); err != nil {
				t.Fatal(err)
			}
			var restored models.SmartGroup
			if err := target.First(&restored, 9).Error; err != nil {
				t.Fatal(err)
			}
			if legacy {
				if restored.CandidateConditions != "" || restored.MemberConditions != "" {
					t.Fatal("legacy backup gained filters")
				}
			} else if restored.CandidateConditions != group.CandidateConditions || restored.MemberConditions != group.MemberConditions {
				t.Fatal("conditions lost in restore")
			}
		})
	}
}
