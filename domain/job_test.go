package domain_test

import (
	"testing"
	"time"

	"github.com/gtkpad/video-encoder/domain"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestNewJob(t *testing.T) {
	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "file-path"
	video.CreatedAt = time.Now()

	job, err := domain.NewJob("output-bucket-path", "status", video)
	require.NotNil(t, job)
	require.Nil(t, err)
}