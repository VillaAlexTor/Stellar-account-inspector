package sentinel

import (
	"testing"

	"github.com/stellar-account-inspector/sentinel-api/internal/horizon"
)

const testAccount = "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func TestEvaluateSensitiveOperations(t *testing.T) {
	baseState := AccountState{
		AccountID:     testAccount,
		MasterWeight:  1,
		LowThreshold:  1,
		MedThreshold:  2,
		HighThreshold: 2,
		Signers: map[string]int{
			testAccount: 1,
			"GBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB": 1,
		},
		KnownIssuers: map[string]bool{},
	}

	tests := []struct {
		name      string
		operation horizon.Operation
		wantRule  string
		severity  string
	}{
		{
			name: "signer added",
			operation: horizon.Operation{
				Type:          "set_options",
				SourceAccount: testAccount,
				SignerKey:     stringPointer("GCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
				SignerWeight:  intPointer(1),
			},
			wantRule: "SIGNER_ADDED",
			severity: "warning",
		},
		{
			name: "signer removed",
			operation: horizon.Operation{
				Type:          "set_options",
				SourceAccount: testAccount,
				SignerKey:     stringPointer("GBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
				SignerWeight:  intPointer(0),
			},
			wantRule: "SIGNER_REMOVED",
			severity: "warning",
		},
		{
			name: "signer weight changed",
			operation: horizon.Operation{
				Type:          "set_options",
				SourceAccount: testAccount,
				SignerKey:     stringPointer("GBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
				SignerWeight:  intPointer(2),
			},
			wantRule: "SIGNER_WEIGHT_CHANGED",
			severity: "warning",
		},
		{
			name: "threshold changed",
			operation: horizon.Operation{
				Type:          "set_options",
				SourceAccount: testAccount,
				HighThreshold: intPointer(3),
			},
			wantRule: "THRESHOLDS_CHANGED",
			severity: "warning",
		},
		{
			name: "new trustline issuer",
			operation: horizon.Operation{
				Type:          "change_trust",
				SourceAccount: testAccount,
				AssetCode:     stringPointer("RISK"),
				AssetIssuer:   stringPointer("GDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"),
				Limit:         stringPointer("1000.0000000"),
			},
			wantRule: "NEW_TRUSTLINE_ISSUER",
			severity: "warning",
		},
		{
			name: "master key zeroed",
			operation: horizon.Operation{
				Type:            "set_options",
				SourceAccount:   testAccount,
				MasterKeyWeight: intPointer(0),
			},
			wantRule: "MASTER_KEY_ZEROED",
			severity: "critical",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alerts := Evaluate(test.operation, baseState)
			if len(alerts) != 1 {
				t.Fatalf("expected one alert, got %d: %#v", len(alerts), alerts)
			}
			if alerts[0].RuleID != test.wantRule || alerts[0].Severity != test.severity {
				t.Fatalf("unexpected alert: %#v", alerts[0])
			}
		})
	}
}

func TestEvaluateIgnoresKnownIssuerAndOtherSource(t *testing.T) {
	issuer := "GEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE"
	state := AccountState{
		AccountID:    testAccount,
		Signers:      map[string]int{testAccount: 1},
		KnownIssuers: map[string]bool{issuer: true},
		MasterWeight: 1,
	}
	operation := horizon.Operation{
		Type:          "change_trust",
		SourceAccount: testAccount,
		AssetIssuer:   &issuer,
		Limit:         stringPointer("100"),
	}
	if alerts := Evaluate(operation, state); len(alerts) != 0 {
		t.Fatalf("known issuer should not alert: %#v", alerts)
	}

	operation.SourceAccount = "GFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	operation.AssetIssuer = stringPointer("G2222222222222222222222222222222222222222222222222222222")
	if alerts := Evaluate(operation, state); len(alerts) != 0 {
		t.Fatalf("operation from another source should not alert: %#v", alerts)
	}
}

func intPointer(value int) *int          { return &value }
func stringPointer(value string) *string { return &value }
