package domain_test

import (
	"testing"
	"time"

	"github.com/gtkpad/video-encoder/domain"
	"github.com/stretchr/testify/require"
)

func TestValidateIfVideoIsEmpty(t *testing.T) {
	video := domain.NewVideo()
	err := video.Validate()

	require.Error(t, err)
}

func TestVideoIdIsNotAUuid(t *testing.T) {
	video := domain.NewVideo()

	video.ID = "9090020f-1b46-41e3-88df-99bf0e6c4536"
	video.ResourceID = "resource-id"
	video.FilePath = "file-path"
	video.CreatedAt = time.Now()

	err := video.Validate()
	require.Nil(t, err)
}
