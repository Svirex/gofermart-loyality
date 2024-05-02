package services

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestAccrualResponseUnmarshal(t *testing.T) {
	j := `
	{
        "order": "367347345",
        "status": "PROCESSED",
        "accrual": 500
    }
	`
	var ar AccrualResponse
	err := json.Unmarshal([]byte(j), &ar)
	require.NoError(t, err)

	require.Equal(t, "367347345", ar.OrderNum)
	require.Equal(t, Processed, ar.Status)
	d := decimal.NewFromFloat(500)
	require.Equal(t, Amount(d), ar.Accrual)
}

func TestAccrualResponseUnmarshalWithoutAccrual(t *testing.T) {
	j := `
	{
        "order": "367347345",
        "status": "PROCESSED"
    }
	`
	var ar AccrualResponse
	err := json.Unmarshal([]byte(j), &ar)
	require.NoError(t, err)

	require.Equal(t, "367347345", ar.OrderNum)
	require.Equal(t, Processed, ar.Status)
	d := decimal.Decimal{}
	require.Equal(t, Amount(d), ar.Accrual)
}

func TestAccrualResponseMarshal(t *testing.T) {
	ar := AccrualResponse{
		OrderNum: "367347345",
		Status:   Processing,
		Accrual:  Amount(decimal.NewFromFloat(0.143123)),
	}
	b, err := json.Marshal(ar)
	require.NoError(t, err)
	expected := `{"order":"367347345","status":"PROCESSING","accrual":0.143123}`
	require.Equal(t, expected, string(b))
}
