package services

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/gtkpad/video-encoder/domain"
	"github.com/gtkpad/video-encoder/framework/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

var Mutex = &sync.Mutex{}

type JobWorkerResult struct {
	Job 			*domain.Job
	Message 	*amqp.Delivery
	Error 		error
}

func JobWorker(messageChannel chan amqp.Delivery, returnChannel chan JobWorkerResult, jobService JobService, job domain.Job, workerId int) {
	
	
	for message := range messageChannel {
		err := utils.IsJson(string(message.Body))
		if err != nil {
			returnChannel <- returnJobResult(&domain.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		jobService.VideoService.Video.ID = uuid.NewV4().String()
		Mutex.Unlock()
		if err != nil {
			returnChannel <- returnJobResult(&domain.Job{}, message, err)
			continue
		}

		err = jobService.VideoService.Video.Validate()
		if err != nil {
			returnChannel <- returnJobResult(&domain.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		err = jobService.VideoService.InsertVideo()
		Mutex.Unlock()
		if err != nil {
			returnChannel <- returnJobResult(&domain.Job{}, message, err)
			continue
		}

		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("outputBucketName")
		job.ID = uuid.NewV4().String()
		job.Status = "STARTING"
		job.CreatedAt = time.Now()

		Mutex.Lock()
		_, err = jobService.JobRepository.Insert(&job)
		Mutex.Unlock()

		if err != nil {
			returnChannel <- returnJobResult(&domain.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		jobService.Job = &job
		Mutex.Unlock()
		err = jobService.Start()
		if err != nil {
			returnChannel <- returnJobResult(&domain.Job{}, message, err)
			continue
		}
		
		returnChannel <- returnJobResult(&job, message, nil)
	}
}

func returnJobResult(job *domain.Job, message amqp.Delivery, error error) JobWorkerResult {
	return JobWorkerResult{
		Job: job,
		Message: &message,
		Error: error,
	}
}