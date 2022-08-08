package services

import (
	"errors"
	"os"
	"strconv"

	"github.com/gtkpad/video-encoder/application/repositories"
	"github.com/gtkpad/video-encoder/domain"
)

type JobService struct {
	Job *domain.Job
	JobRepository repositories.JobRepository
	VideoService VideoService
}

func (service *JobService) Start() error {
	err := service.changeJobStatus("DOWNLOADING")

	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Download(os.Getenv("inputBucketName"))

	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("FRAGMENTING")

	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Fragment()	

	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("ENCODING")

	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Encode()

	if err != nil {
		return service.failJob(err)
	}

	err = service.performUpload()

	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("FINISHING")

	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Finish()

	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("COMPLETED")

	if err != nil {
		return service.failJob(err)
	}

	return nil
}

func (service *JobService) performUpload() error {
	err := service.changeJobStatus("UPLOADING")

	if err != nil {
		return service.failJob(err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("outputBucketName")
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + service.VideoService.Video.ID
	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))
	doneUpload := make(chan string)

	go videoUpload.ProcessUpload(concurrency, doneUpload)

	var uploadResult string
	uploadResult = <-doneUpload

	if uploadResult != "upload completed" {
		return service.failJob(errors.New(uploadResult))
	}

	return err
}

func (service *JobService) changeJobStatus(status string) error {
	var err error
	Mutex.Lock()
	service.Job.Status = status
	service.Job, err = service.JobRepository.Update(service.Job)
	Mutex.Lock()
	if err != nil {
		return service.failJob(err)
	}
	return nil
}

func (service *JobService) failJob(error error) error {

	service.Job.Status = "FAILED"
	service.Job.Error = error.Error()
	_, err := service.JobRepository.Update(service.Job)

	if err != nil {
		return err
	}

	return error
}