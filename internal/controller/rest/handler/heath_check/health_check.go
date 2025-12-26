package heath_check

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func Handler(usecase usecase.HealthCheckUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		result, err := usecase.CheckSystemConnections(request.Context())
		if err != nil || !result {
			writer.WriteHeader(http.StatusInternalServerError)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
	}
}
