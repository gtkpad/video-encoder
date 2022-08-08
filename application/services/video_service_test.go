package services_test

import (
	"log"
	"testing"
	"time"

	"github.com/gtkpad/video-encoder/application/repositories"
	"github.com/gtkpad/video-encoder/application/services"
	"github.com/gtkpad/video-encoder/domain"
	"github.com/gtkpad/video-encoder/framework/database"
	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func prepare() (*domain.Video, repositories.VideoRepositoryDb) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "testing.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}
	
	return video, repo
}

func TestVideoServiceDownload(t *testing.T) {
	video, repo := prepare()
	repo.Insert(video)

	service := services.NewVideoService()
	service.Video = video
	service.VideoRepository = repo

	err := service.Download("testing-video-encoder")
	// log.Printf("%v", err.Error())
	require.Nil(t, err)

	err = service.Fragment()
	require.Nil(t, err)

	err = service.Encode()
	require.Nil(t, err)

	err = service.Finish()
	require.Nil(t, err)
}