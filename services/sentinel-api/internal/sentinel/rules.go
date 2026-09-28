package sentinel

import (
	"fmt"
	"strings"

	"github.com/stellar-account-inspector/sentinel-api/internal/horizon"
)

type AlertDraft struct {
	RuleID   string
	Severity string
	Message  string
}

type Rule func(horizon.Operation, AccountState) *AlertDraft

var Rules = []Rule{
	findMasterKeyZeroed,
	findSignerChange,
	findThresholdChange,
	findNewTrustlineIssuer,
}

func Evaluate(operation horizon.Operation, state AccountState) []AlertDraft {
	if operation.SourceAccount != "" && operation.SourceAccount != state.AccountID {
		return nil
	}
	alerts := make([]AlertDraft, 0, len(Rules))
	for _, rule := range Rules {
		if alert := rule(operation, state); alert != nil {
			alerts = append(alerts, *alert)
		}
	}
	return alerts
}

func findMasterKeyZeroed(operation horizon.Operation, state AccountState) *AlertDraft {
	if operation.Type != "set_options" || operation.MasterKeyWeight == nil {
		return nil
	}
	if *operation.MasterKeyWeight != 0 || state.MasterWeight == 0 {
		return nil
	}
	return &AlertDraft{
		RuleID:   "MASTER_KEY_ZEROED",
		Severity: "critical",
		Message:  "La master key redujo su peso a 0. Verifica de inmediato que existan firmantes alternativos con peso suficiente antes de que la cuenta pierda control operativo.",
	}
}

func findSignerChange(operation horizon.Operation, state AccountState) *AlertDraft {
	if operation.Type != "set_options" || operation.SignerKey == nil || operation.SignerWeight == nil {
		return nil
	}
	key := *operation.SignerKey
	weight := *operation.SignerWeight
	previousWeight, existed := state.Signers[key]
	shortKey := abbreviateKey(key)

	switch {
	case !existed && weight > 0:
		return &AlertDraft{
			RuleID:   "SIGNER_ADDED",
			Severity: "warning",
			Message:  fmt.Sprintf("Se agregó el firmante %s con peso %d. Confirma que la clave pertenece a un participante autorizado.", shortKey, weight),
		}
	case existed && weight == 0:
		return &AlertDraft{
			RuleID:   "SIGNER_REMOVED",
			Severity: "warning",
			Message:  fmt.Sprintf("Se removió el firmante %s, que tenía peso %d. Revisa si la capacidad multifirma sigue alcanzando los umbrales.", shortKey, previousWeight),
		}
	case existed && weight != previousWeight:
		return &AlertDraft{
			RuleID:   "SIGNER_WEIGHT_CHANGED",
			Severity: "warning",
			Message:  fmt.Sprintf("El firmante %s cambió de peso %d a %d. El poder relativo de autorización de la cuenta fue modificado.", shortKey, previousWeight, weight),
		}
	default:
		return nil
	}
}

func findThresholdChange(operation horizon.Operation, state AccountState) *AlertDraft {
	if operation.Type != "set_options" {
		return nil
	}
	changes := make([]string, 0, 3)
	if operation.LowThreshold != nil && *operation.LowThreshold != state.LowThreshold {
		changes = append(changes, fmt.Sprintf("bajo %d→%d", state.LowThreshold, *operation.LowThreshold))
	}
	if operation.MedThreshold != nil && *operation.MedThreshold != state.MedThreshold {
		changes = append(changes, fmt.Sprintf("medio %d→%d", state.MedThreshold, *operation.MedThreshold))
	}
	if operation.HighThreshold != nil && *operation.HighThreshold != state.HighThreshold {
		changes = append(changes, fmt.Sprintf("alto %d→%d", state.HighThreshold, *operation.HighThreshold))
	}
	if len(changes) == 0 {
		return nil
	}
	return &AlertDraft{
		RuleID:   "THRESHOLDS_CHANGED",
		Severity: "warning",
		Message:  fmt.Sprintf("Se modificaron los umbrales de autorización: %s. Comprueba que los pesos actuales todavía permiten operar sin debilitar el control.", strings.Join(changes, ", ")),
	}
}

func findNewTrustlineIssuer(operation horizon.Operation, state AccountState) *AlertDraft {
	if operation.Type != "change_trust" || operation.AssetIssuer == nil || !operation.HasPositiveTrustLimit() {
		return nil
	}
	issuer := *operation.AssetIssuer
	if issuer == "" || state.KnownIssuers[issuer] {
		return nil
	}
	asset := "activo sin código"
	if operation.AssetCode != nil && *operation.AssetCode != "" {
		asset = *operation.AssetCode
	}
	return &AlertDraft{
		RuleID:   "NEW_TRUSTLINE_ISSUER",
		Severity: "warning",
		Message:  fmt.Sprintf("Se creó una trustline para %s hacia el emisor no visto %s. Verifica la identidad y las políticas del emisor antes de recibir fondos.", asset, abbreviateKey(issuer)),
	}
}

func abbreviateKey(key string) string {
	if len(key) <= 14 {
		return key
	}
	return key[:7] + "…" + key[len(key)-6:]
}
