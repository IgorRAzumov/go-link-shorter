package healthcheck

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

// Handler возвращает обработчик GET /ping — проверка доступности БД.
//
// @Summary      Health check
// @Description  Проверка доступности базы данных
// @Tags         health
// @Success      200  {string}  string  "OK"
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /ping [get]
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
