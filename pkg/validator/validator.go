package validator

import (
	"mime/multipart"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"template-golang/pkg/apperror"
	"template-golang/pkg/logger"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Gunakan tag `json` sebagai nama field di error
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	// Default validation rules yang berguna

	// Contoh: password minimal 8 karakter, harus ada huruf & angka
	validate.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		p := fl.Field().String()
		if len(p) < 8 {
			return false
		}
		hasLetter := false
		hasNumber := false
		for _, ch := range p {
			switch {
			case ch >= '0' && ch <= '9':
				hasNumber = true
			case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
				hasLetter = true
			}
		}
		return hasLetter && hasNumber
	})

	// Contoh: username harus alfanumerik
	validate.RegisterValidation("alphanum_space", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		for _, ch := range v {
			if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == ' ') {
				return false
			}
		}
		return true
	})

	// Validate latitude (-90 to 90)
	validate.RegisterValidation("latitude", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
			v := fl.Field().Float()
			return v >= -90 && v <= 90
		}
		return false
	})

	// Validate longitude (-180 to 180)
	validate.RegisterValidation("longitude", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
			v := fl.Field().Float()
			return v >= -180 && v <= 180
		}
		return false
	})

	// Contoh: positive number
	validate.RegisterValidation("positive", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int() > 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return field.Uint() > 0
		default:
			return false
		}
	})

	// Validate image file types (jpg, jpeg, png)
	validate.RegisterValidation("image", func(fl validator.FieldLevel) bool {
		v, ok := fl.Field().Interface().(multipart.FileHeader)
		if !ok {
			return false
		}

		contentType := v.Header.Get("Content-Type")
		validTypes := []string{"image/jpg", "image/jpeg", "image/png"}

		for _, mimeType := range validTypes {
			if contentType == mimeType {
				return true
			}
		}
		return false
	})

	// Validate image file size with custom max size in MB
	validate.RegisterValidation("size", func(fl validator.FieldLevel) bool {
		v := fl.Field().Int()
		param := fl.Param()
		maxSizeMB, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return false
		}
		maxSizeBytes := maxSizeMB * 1024 * 1024 // Convert MB to bytes
		return v <= maxSizeBytes
	})

	// Validate image file types (jpg, jpeg, png)
	validate.RegisterValidation("image", func(fl validator.FieldLevel) bool {
		v := strings.ToLower(fl.Field().String())
		validTypes := []string{".jpg", ".jpeg", ".png"}
		for _, ext := range validTypes {
			if strings.HasSuffix(v, ext) {
				return true
			}
		}
		return false
	})

	// Validate image file size with custom max size in MB
	validate.RegisterValidation("size", func(fl validator.FieldLevel) bool {
		v := fl.Field().Int()
		param := fl.Param()
		maxSizeMB, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return false
		}
		maxSizeBytes := maxSizeMB * 1024 * 1024 // Convert MB to bytes
		return v <= maxSizeBytes
	})
}

// ValidateStruct pakai validator global
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		errs := make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			message := ""
			switch e.Tag() {
			case "required":
				message = e.Field() + " is required"
			case "email":
				message = e.Field() + " must be a valid email address"
			case "min":
				message = e.Field() + " must be at least " + e.Param() + " characters long"
			case "max":
				message = e.Field() + " must not exceed " + e.Param() + " characters"
			case "strong_password":
				message = e.Field() + " must be at least 8 characters and contain both letters and numbers"
			case "alphanum_space":
				message = e.Field() + " must only contain letters, numbers and spaces"
			case "latitude":
				message = e.Field() + " must be a valid latitude between -90 and 90"
			case "longitude":
				message = e.Field() + " must be a valid longitude between -180 and 180"
			case "positive":
				message = e.Field() + " must be a positive number"
			case "image":
				message = e.Field() + " must be a valid image file (jpg, jpeg, png)"
			case "size":
				message = e.Field() + " must not exceed " + e.Param() + " MB"
			default:
				message = e.Field() + " failed " + e.Tag() + " validation"
			}
			errs[strings.ToLower(e.Field())] = message
		}
		return apperror.New(
			"VALIDATION_ERROR",
			"Validation Failed",
			http.StatusUnprocessableEntity,
			errs,
			string(debug.Stack()),
		)
	}
	return nil
}

// StringToInt converts string to int with error handling
func StringToInt(str string) (int, error) {
	if str == "" {
		return 0, nil
	}
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, apperror.New(
			"INVALID_INT",
			"Invalid integer value",
			http.StatusBadRequest,
			nil,
			string(debug.Stack()),
		)
	}
	logger.L().Info(num)
	return num, nil
}
