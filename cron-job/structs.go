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
	Unix   int64 `json:"date"`
	Status int   `json:"status"`
}

var statusMap = map[int]string{
	0: "Unknown / not executed yet",
	1: "OK",
	2: "Failed (DNS error)",
	3: "Failed (could not connect to host)",
	4: "Failed (HTTP error)",
	5: "Failed (timeout)",
	6: "Failed (too much response data)",
	7: "Failed (invalid URL)",
	8: "Failed (internal errors)",
	9: "Failed (unknown reason)",
}

func (jh JobHistory) GetStatusString() string {
	return statusMap[jh.Status]
}

type JobData struct {
	ID            int          `json:"jobId,omitempty"`
	Title         string       `json:"title,omitempty"`
	URL           string       `json:"url,omitempty"`
	RequestMethod int          `json:"requestMethod,omitempty"`
	Enabled       bool         `json:"enabled,omitempty"`
	Schedule      Schedule     `json:"schedule,omitempty"`
	ExtendedData  ExtendedData `json:"extendedData,omitempty"`
	SaveResponse  bool         `json:"saveResponses,omitempty"`
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

// implement encoding.BinaryMarshaler & encoding.BinaryUnmarshaler interface
type JobReqData struct {
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
	JobID             int    `json:"job_id,omitempty"`
}

func (b JobReqData) MarshalBinary() ([]byte, error) {
	return json.Marshal(b)
}

func (b *JobReqData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, b)
}

func (b JobReqData) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

func (b JobReqData) Eq(b1 JobReqData) bool {
	return b.AccessToken == b1.AccessToken && b.AccessTokenSecret == b1.AccessTokenSecret
}

type JobHistories []JobHistory

func (h JobHistories) MarshalBinary() ([]byte, error) {
	return json.Marshal(h)
}

func (h *JobHistories) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, h)
}
