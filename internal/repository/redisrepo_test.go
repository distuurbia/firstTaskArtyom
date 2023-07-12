package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetCache(t *testing.T) {
	err := rdsRps.SetCache(context.Background(), &testModel)
	require.NoError(t, err)
}

func TestGetCache(t *testing.T) {
	getCar, err := rdsRps.GetCache(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, testModel.ID, getCar.ID)
	require.Equal(t, testModel.Brand, getCar.Brand)
	require.Equal(t, testModel.ProductionYear, getCar.ProductionYear)
	require.Equal(t, testModel.IsRunning, getCar.IsRunning)
}

func TestDeleteCache(t *testing.T) {
	err := rdsRps.DeleteCache(context.Background(), testModel.ID)
	require.NoError(t, err)
}
