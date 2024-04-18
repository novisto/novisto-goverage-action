package httperrors

import "github.com/labstack/echo/v4"

type GoverageError struct {
	Message string `json:"message"`
}

func WriteResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, &GoverageError{Message: message})
}
