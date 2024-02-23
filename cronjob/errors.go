package cronjob

import "errors"

var (
	ErrFetchHistory     = errors.New("something went wrong while trying to fetch job history data")
	ErrIDKHowToNameThis = errors.New("something went wrong")
)
