package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/regen-network/regen-ledger/x/ecocredit"
	"github.com/regen-network/regen-ledger/x/ecocredit/simulation"
	"github.com/stretchr/testify/require"
)

func TestParamChanges(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	bz, err := json.Marshal([]*ecocredit.CreditType{
		{
			Name:         "carbon",
			Abbreviation: "C",
			Unit:         "metric ton CO2 equivalent",
			Precision:    6,
		},
		{
			Name:         "biodiversity",
			Abbreviation: "BIO",
			Unit:         "ton",
			Precision:    6,
		}},
	)
	require.NoError(t, err)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"ecocredit/CreditClassFee", "CreditClassFee", "[{\"denom\":\"stake\",\"amount\":\"1\"}]", "ecocredit"},
		{"ecocredit/AllowlistEnabled", "AllowlistEnabled", "true", "ecocredit"},
		{"ecocredit/AllowedClassCreators", "AllowedClassCreators", "[\"cosmos18wa8fq26625ap562yvxvd026gtm3yq904v3kk5\",\"cosmos1xrstn0emdpfkh3ajhmwgmxf4cj60zem2th7cy7\"]", "ecocredit"},
		{"ecocredit/CreditTypes", "CreditTypes", string(bz), "ecocredit"},
	}

	paramChanges := simulation.ParamChanges(r)

	require.Len(t, paramChanges, 4)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}