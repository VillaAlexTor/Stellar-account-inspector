package horizon

import (
	"encoding/json"
	"strconv"
)

type Signer struct {
	Key    string `json:"key"`
	Weight int    `json:"weight"`
	Type   string `json:"type"`
}

type Thresholds struct {
	Low  int `json:"low_threshold"`
	Med  int `json:"med_threshold"`
	High int `json:"high_threshold"`
}

type Balance struct {
	AssetType   string `json:"asset_type"`
	AssetIssuer string `json:"asset_issuer"`
}

type AccountSnapshot struct {
	AccountID  string     `json:"account_id"`
	Thresholds Thresholds `json:"thresholds"`
	Signers    []Signer   `json:"signers"`
	Balances   []Balance  `json:"balances"`
}

type Operation struct {
	ID              string          `json:"id"`
	PagingToken     string          `json:"paging_token"`
	Type            string          `json:"type"`
	SourceAccount   string          `json:"source_account"`
	CreatedAt       string          `json:"created_at"`
	SignerKey       *string         `json:"signer_key"`
	SignerWeight    *int            `json:"signer_weight"`
	MasterKeyWeight *int            `json:"master_key_weight"`
	LowThreshold    *int            `json:"low_threshold"`
	MedThreshold    *int            `json:"med_threshold"`
	HighThreshold   *int            `json:"high_threshold"`
	AssetIssuer     *string         `json:"asset_issuer"`
	AssetCode       *string         `json:"asset_code"`
	Limit           *string         `json:"limit"`
	Raw             json.RawMessage `json:"-"`
}

func (operation Operation) HasPositiveTrustLimit() bool {
	if operation.Limit == nil {
		return true
	}
	value, err := strconv.ParseFloat(*operation.Limit, 64)
	return err != nil || value > 0
}
