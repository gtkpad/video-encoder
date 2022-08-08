package utils_test

import (
	"testing"

	"github.com/gtkpad/video-encoder/framework/utils"
	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	json := `{
		"name": "John Doe",
		"age": 30
	}`

	err := utils.IsJson(json)

	require.Nil(t, err)

	json = `{asdasdsaçk`

	err = utils.IsJson(json)

	require.NotNil(t, err)
}