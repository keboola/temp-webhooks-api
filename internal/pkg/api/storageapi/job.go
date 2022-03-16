package storageapi

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/spf13/cast"
)

func (a *Api) GetJob(jobId int) (*model.Job, error) {
	response := a.GetJobRequest(jobId).Send().Response
	if response.HasResult() {
		return response.Result().(*model.Job), nil
	}
	return nil, response.Err()
}

// GetJobRequest https://keboola.docs.apiary.io/#reference/jobs/manage-jobs/job-detail
func (a *Api) GetJobRequest(jobId int) *client.Request {
	job := &model.Job{}
	return a.
		NewRequest(resty.MethodGet, "jobs/{jobId}").
		SetPathParam("jobId", cast.ToString(jobId)).
		SetResult(job)
}

// nolint: unused
func waitForJob(a *Api, parentRequest *client.Request, job *model.Job, onJobSuccess client.ResponseCallback) client.ResponseCallback {
	// Check job
	backoff := newBackoff()
	var checkJobStatus client.ResponseCallback
	checkJobStatus = func(response *client.Response) {
		// Check status
		if job.Status == "success" {
			if onJobSuccess != nil {
				onJobSuccess(response)
			}
			return
		} else if job.Status == "error" {
			err := fmt.Errorf("job failed: %v", job.Results)
			response.SetErr(err)
			return
		}

		// Wait and check again
		delay := backoff.NextBackOff()
		if delay == backoff.Stop {
			err := fmt.Errorf("timeout: timeout while waiting for the storage job to complete")
			response.SetErr(err)
			return
		}

		// Try again
		request := a.
			GetJobRequest(job.Id).
			SetResult(job).
			OnSuccess(checkJobStatus)

		parentRequest.WaitFor(request)
		time.Sleep(delay)
		response.Sender().Request(request).Send()
	}
	return checkJobStatus
}
