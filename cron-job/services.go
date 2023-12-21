package cronjob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type CronJobService interface {
	CreateOrUpdate(string, string, string) (int, error)
	GetHistory(int) ([]JobHistory, error)
}

type cronjobService struct {
	apiKey string
	client *http.Client
	logger *slog.Logger
}

func NewCronJobService(APIKey string, client *http.Client, logger *slog.Logger) cronjobService {
	return cronjobService{APIKey, client, logger}
}

func (c cronjobService) CreateOrUpdate(title, accessToken, accessTokenSecret string) (int, error) {
	var (
		id     int
		exists bool
	)
	jobs, err := c.getAll()
	if err != nil {
		c.logger.Error(err.Error())
		return -1, err
	}

	contains := func(jobs []JobData, title string) (int, bool) {
		for _, job := range jobs {
			if job.Title == title {
				return job.ID, true
			}
		}
		return -1, false
	}

	if id, exists = contains(jobs, title); exists {
		err = c.update(id, accessToken, accessTokenSecret)
	} else {
		id, err = c.create(title, accessToken, accessTokenSecret)
	}

	if err != nil {
		c.logger.Error(err.Error())
		return -1, err
	}

	return id, nil
}

func (c cronjobService) GetHistory(id int) ([]JobHistory, error) {
	var historyResp GetHistoryResp

	res, err := c.client.Do(c.newGetHistoryRequest(id))
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		c.logger.Info("unable to fetch job history", "Code", res.StatusCode)
		return nil, ErrFetchHistory
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&historyResp); err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	return historyResp.History, nil
}

func (c cronjobService) create(title, accessToken, accessTokenSecret string) (int, error) {
	var job JobData
	res, err := c.client.Do(c.newCreateRequest(title, accessToken, accessTokenSecret))
	if err != nil {
		return -1, err
	}

	if res.StatusCode != http.StatusOK {
		c.logger.Info("unable to create a new job", "Code", res.StatusCode)
		return -1, ErrIDKHowToNameThis
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return -1, err
	}

	return job.ID, nil
}

func (c cronjobService) update(id int, accessToken, accessTokenSecret string) error {
	res, err := c.client.Do(c.newUpdateRequest(id, accessToken, accessTokenSecret))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		c.logger.Info("unable to update job data", "Code", res.StatusCode)
		return ErrIDKHowToNameThis
	}
	return nil
}

func (c cronjobService) getAll() ([]JobData, error) {
	var respObj GetAllResp
	res, err := c.client.Do(c.newGetAllRequest())
	if err != nil {
		return respObj.Jobs, err
	}

	if res.StatusCode != http.StatusOK {
		c.logger.Info("unable to fetch all job", "Code", res.StatusCode)
		return respObj.Jobs, ErrIDKHowToNameThis
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&respObj); err != nil {
		return respObj.Jobs, err
	}

	return respObj.Jobs, nil
}

func (c cronjobService) newCreateRequest(title, at, ats string) *http.Request {
	extendedBody := JobReqBody{at, ats}
	reqBody := CreateReqBody{
		Job: JobData{
			Title:         title,
			URL:           "https://twitter-moon.vercel.app/picture",
			RequestMethod: 1,
			Enabled:       true,
			Schedule: Schedule{
				Timezone:  "Asia/Jakarta",
				ExpiresAt: 0,
				Hours:     []int{-1},
				MDays:     []int{-1},
				Minutes:   []int{31},
				Months:    []int{-1},
				WDays:     []int{-1},
			},
			ExtendedData: ExtendedData{
				Headers: map[string]any{
					"Content-Type": "application/json",
				},
				Body: extendedBody.String(),
			},
		},
	}

	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "https://api.cron-job.org/jobs", &buf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	return req
}

func (c cronjobService) newGetAllRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "https://api.cron-job.org/jobs", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	return req
}

func (c cronjobService) newUpdateRequest(id int, at, ats string) *http.Request {
	reqBody := CreateReqBody{
		Job: JobData{
			ExtendedData: ExtendedData{
				Body: JobReqBody{
					AccessToken:       at,
					AccessTokenSecret: ats,
				}.String(),
			},
		},
	}
	var buf bytes.Buffer

	_ = json.NewEncoder(&buf).Encode(reqBody)
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://api.cron-job.org/jobs/%d", id), &buf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	return req
}

func (c cronjobService) newGetHistoryRequest(id int) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.cron-job.org/jobs/%d/history", id), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	return req
}
