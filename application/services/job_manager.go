package services

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/gtkpad/video-encoder/application/repositories"
	"github.com/gtkpad/video-encoder/domain"
	"github.com/gtkpad/video-encoder/framework/queue"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type JobManager struct {
	Db 								*gorm.DB
	Domain 						domain.Job
	MessageChannel 		chan amqp.Delivery
	JobReturnChannel 	chan JobWorkerResult
	RabbitMQ 					*queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error 	string `json:"error"`
}

func NewJobManager(db *gorm.DB, rabbitMQ *queue.RabbitMQ, jobReturnChannel chan JobWorkerResult, messageChannel chan amqp.Delivery) *JobManager {
	return &JobManager{
		Db: db,
		RabbitMQ: rabbitMQ,
		JobReturnChannel: jobReturnChannel,
		MessageChannel: messageChannel,
		Domain: domain.Job{},
	}
}


func (jobManager *JobManager) Start(channel *amqp.Channel) {
	videoService := NewVideoService()
	videoService.VideoRepository = repositories.VideoRepositoryDb{Db: jobManager.Db}

	JobService := JobService{
		JobRepository: repositories.JobRepositoryDb{Db: jobManager.Db},
		VideoService: videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))
	
	if err != nil {
		log.Fatalf("error loading var: CONCURRENCY_WORKERS: %s\n", err)
	}

	for qtdProcesses := 0; qtdProcesses < concurrency; qtdProcesses++ {
		go JobWorker(jobManager.MessageChannel, jobManager.JobReturnChannel, JobService, jobManager.Domain, qtdProcesses)
	}

	for jobResult := range jobManager.JobReturnChannel {
		if jobResult.Error != nil {
			err = jobManager.checkParseErrors(jobResult)
		} else {
			err = jobManager.notifySuccess(jobResult, channel)
		}

		if err != nil {
			jobResult.Message.Reject(false)
		}
	}

}

func (JobManager *JobManager) notifySuccess(jobResult JobWorkerResult, channel *amqp.Channel) error {
	Mutex.Lock()
	jobJson, err := json.Marshal(jobResult.Job)
	Mutex.Unlock()

	if err != nil {
		return err
	}

	err = JobManager.notify(jobJson)

	if err != nil {
		return err
	}

	err = jobResult.Message.Ack(false)

	if err != nil {
		return err
	}

	return nil
}

func (jobManager *JobManager) checkParseErrors(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		log.Printf("MessageID: %v. Error during job: %v with video: %v. Error: %v",
			jobResult.Message.DeliveryTag, jobResult.Job.ID, jobResult.Job.Video.ID, jobResult.Error.Error())
	} else {
		log.Printf("MessageID: %v. Error parsing message: %v", jobResult.Message.DeliveryTag, jobResult.Error.Error())
	}

	errorMsg := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error: jobResult.Error.Error(),
	}

	jobJson, err := json.Marshal(errorMsg)

	if err != nil {
		return err
	}

	err = jobManager.notify(jobJson)

	if err != nil {
		return err
	}

	err = jobResult.Message.Reject(false)
	
	if err != nil {
		return err
	}

	return nil
}

func (jobManager *JobManager) notify(jobJson []byte) error {
	err := jobManager.RabbitMQ.Notify(
		string(jobJson), 
		"application/json", 
		os.Getenv("RABBITMQ_NOTIFICATION_EX"), 
		os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"),
	)

	if err != nil {
		return err
	}

	return nil
}