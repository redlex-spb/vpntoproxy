package responses

type RespData struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

func OutputSuccessData(data interface{}) *RespData {
	return &RespData{
		Status: true,
		Data:   data,
	}
}

func OutputErrorData(err error) *RespData {
	return &RespData{
		Status: false,
		Data:   err.Error(),
	}
}
