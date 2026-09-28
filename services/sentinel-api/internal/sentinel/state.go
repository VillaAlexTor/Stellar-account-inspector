package sentinel

import (
	"encoding/json"

	"github.com/stellar-account-inspector/sentinel-api/internal/horizon"
)

type AccountState struct {
	AccountID    string         `json:"accountId"`
	MasterWeight int            `json:"masterWeight"`
	LowThreshold int            `json:"lowThreshold"`
	MedThreshold int            `json:"medThreshold"`
	HighThreshold int           `json:"highThreshold"`
	Signers      map[string]int `json:"signers"`
	KnownIssuers map[string]bool `json:"knownIssuers"`
}

func StateFromSnapshot(snapshot horizon.AccountSnapshot) AccountState {
	state := AccountState{
		AccountID:     snapshot.AccountID,
		LowThreshold:  snapshot.Thresholds.Low,
		MedThreshold:  snapshot.Thresholds.Med,
		HighThreshold: snapshot.Thresholds.High,
		Signers:       make(map[string]int, len(snapshot.Signers)),
		KnownIssuers:  make(map[string]bool),
	}
	for _, signer := range snapshot.Signers {
		state.Signers[signer.Key] = signer.Weight
		if signer.Key == snapshot.AccountID {
			state.MasterWeight = signer.Weight
		}
	}
	for _, balance := range snapshot.Balances {
		if balance.AssetType != "native" && balance.AssetIssuer != "" {
			state.KnownIssuers[balance.AssetIssuer] = true
		}
	}
	return state
}

func StateFromJSON(value string) (AccountState, error) {
	var state AccountState
	if err := json.Unmarshal([]byte(value), &state); err != nil {
		return AccountState{}, err
	}
	if state.Signers == nil {
		state.Signers = make(map[string]int)
	}
	if state.KnownIssuers == nil {
		state.KnownIssuers = make(map[string]bool)
	}
	return state, nil
}

func (state AccountState) JSON() (string, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func (state AccountState) Clone() AccountState {
	clone := state
	clone.Signers = make(map[string]int, len(state.Signers))
	for key, weight := range state.Signers {
		clone.Signers[key] = weight
	}
	clone.KnownIssuers = make(map[string]bool, len(state.KnownIssuers))
	for issuer, known := range state.KnownIssuers {
		clone.KnownIssuers[issuer] = known
	}
	return clone
}

func (state *AccountState) Apply(operation horizon.Operation) {
	if operation.SourceAccount != "" && operation.SourceAccount != state.AccountID {
		return
	}
	switch operation.Type {
	case "set_options":
		if operation.MasterKeyWeight != nil {
			state.MasterWeight = *operation.MasterKeyWeight
			state.Signers[state.AccountID] = *operation.MasterKeyWeight
		}
		if operation.LowThreshold != nil {
			state.LowThreshold = *operation.LowThreshold
		}
		if operation.MedThreshold != nil {
			state.MedThreshold = *operation.MedThreshold
		}
		if operation.HighThreshold != nil {
			state.HighThreshold = *operation.HighThreshold
		}
		if operation.SignerKey != nil && operation.SignerWeight != nil {
			if *operation.SignerWeight == 0 {
				delete(state.Signers, *operation.SignerKey)
			} else {
				state.Signers[*operation.SignerKey] = *operation.SignerWeight
			}
		}
	case "change_trust":
		if operation.AssetIssuer != nil && *operation.AssetIssuer != "" && operation.HasPositiveTrustLimit() {
			state.KnownIssuers[*operation.AssetIssuer] = true
		}
	}
}
