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
	Body        any `json:"body"`
	APIResponse *APIResponse
	Identifier  string `json:"identifier"`
	Status      Status `json:"status"`
	Unix        int64  `json:"date"`
	HttpStatus  int    `json:"httpStatus"`
}

type JobHistoryDetails struct {
	JobHistoryDetails JobHistory `json:"jobHistoryDetails"`
}

type APIResponse struct {
	Message string `json:"message"`
}

type Status int

const (
	Unknown Status = iota
	OK
	FailedDNSErr
	FailedCouldnotConnect
	FailedHTTPError
	FailedTimeout
	FailedTooMuchData
	FailedInvalidURL
	FailedInternalErr
	FailedUnknown
)

var statusMap = map[Status]string{
	Unknown:               "Unknown / not executed yet",
	OK:                    "OK",
	FailedDNSErr:          "Failed (DNS error)",
	FailedCouldnotConnect: "Failed (could not connect to host)",
	FailedHTTPError:       "Failed (HTTP error)",
	FailedTimeout:         "Failed (timeout)",
	FailedTooMuchData:     "Failed (too much response data)",
	FailedInvalidURL:      "Failed (invalid URL)",
	FailedInternalErr:     "Failed (internal errors)",
	FailedUnknown:         "Failed (unknown reason)",
}

func (jh JobHistory) GetStatusString() string {
	return statusMap[jh.Status]
}

type JobData struct {
	URL           string       `json:"url,omitempty"`
	Title         string       `json:"title,omitempty"`
	ExtendedData  ExtendedData `json:"extendedData,omitempty"`
	Schedule      Schedule     `json:"schedule,omitempty"`
	ID            int          `json:"jobId,omitempty"`
	RequestMethod int          `json:"requestMethod,omitempty"`
	Enabled       bool         `json:"enabled,omitempty"`
	SaveResponse  bool         `json:"saveResponses,omitempty"`
}

type Schedule struct {
	Timezone  string `json:"timezone,omitempty"`
	Hours     []int  `json:"hours,omitempty"`
	MDays     []int  `json:"mdays,omitempty"`
	Minutes   []int  `json:"minutes,omitempty"`
	Months    []int  `json:"months,omitempty"`
	WDays     []int  `json:"wdays,omitempty"`
	ExpiresAt int    `json:"expiresAt,omitempty"`
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
