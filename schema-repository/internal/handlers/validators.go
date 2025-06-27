package handlers

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func RegisterCustomValidators() {
	v := binding.Validator.Engine().(*validator.Validate)

	err := v.RegisterValidation("json_schema", validateJSONSchema)
	if err != nil {
		panic(err)
	}
}

func validateJSONSchema(fl validator.FieldLevel) bool {
	schema := fl.Field().Bytes()

	if _, err := jsonschema.CompileString("", string(schema)); err != nil {
		return false
	}

	return true
}
