package validator

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Create a go-playground/validator, but wrap it in our generic Validator interface
// as our usage is a small slice of the full go-playground/validator's capabilities.
func New() Validator {
	validate := validator.New()
	en := en.New()
	uni := ut.New(en, en)
	transEn, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, transEn)

	return &playgroundValidator{
		validate:            validate,
		universalTranslator: uni,
		englishTranslator:   transEn,
	}
}

type playgroundValidator struct {
	validate            *validator.Validate
	universalTranslator *ut.UniversalTranslator
	englishTranslator   ut.Translator
}

func (g *playgroundValidator) Validate(inputStruct interface{}) ([]string, error) {
	messages := []string{}
	err := g.validate.Struct(inputStruct)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			messages = append(messages, err.Translate(g.englishTranslator))
		}
	}
	return messages, err
}
