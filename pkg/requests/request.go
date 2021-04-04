package requests

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"
)

type CreateVPNParams struct {
	Path string `form:"path" json:"path" binding:"required"`
}

func (params *CreateVPNParams) Bind(r *http.Request) error {
	return validation.ValidateStruct(params,
		validation.Field(&params.Path, validation.Required))
}
