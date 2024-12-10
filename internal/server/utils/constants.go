package utils

type Constants struct {
	PageSize int
}

var AppConstants Constants

func init() {
	AppConstants = Constants{
		PageSize: 15,
	}
}
