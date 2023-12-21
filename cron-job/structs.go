package cronjob

import "encoding/json"

type CreateReqBody struct {
	Job JobData `json:"job"`
}

type GetAllResp struct {
	Jobs []JobData `json:"jobs"`
}

type GetHistoryResp struct {
	History []JobHistory `json:"history"`
}

type JobHistory struct {
	Unix       int64  `json:"date"`
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
}

type JobData struct {
	ID            int          `json:"jobId,omitempty"`
	Title         string       `json:"title,omitempty"`
	URL           string       `json:"url,omitempty"`
	RequestMethod int          `json:"requestMethod,omitempty"`
	Enabled       bool         `json:"enabled,omitempty"`
	Schedule      Schedule     `json:"schedule,omitempty"`
	ExtendedData  ExtendedData `json:"extendedData,omitempty"`
}

type Schedule struct {
	Timezone  string `json:"timezone,omitempty"`
	ExpiresAt int    `json:"expiresAt,omitemty"`
	Hours     []int  `json:"hours,omitempty"`
	MDays     []int  `json:"mdays,omitempty"`
	Minutes   []int  `json:"minutes,omitempty"`
	Months    []int  `json:"months,omitempty"`
	WDays     []int  `json:"wdays,omitempty"`
}

type ExtendedData struct {
	Headers map[string]any `json:"headers,omitempty"`
	Body    string         `json:"body,omitempty"`
}

type JobReqBody struct {
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

func (b JobReqBody) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}
