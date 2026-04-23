package stats

import (
	"encoding/json"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

// StatisticHandler возвращает обработчик GET /api/internal/stats.
//
// @Summary      Service statistics
// @Description  Возвращает количество сокращённых URL и пользователей сервиса
// @Tags         internal
// @Produce      json
// @Success      200  {object}  model.ServiceStats
// @Failure      403  {string}  string  "Forbidden"
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/internal/stats [get]
func StatisticHandler(statsUsecase usecase.StatsUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var stats model.ServiceStats
		var err error
		stats, err = statsUsecase.GetStats(request.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to get stats")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(writer).Encode(stats); err != nil {
			log.Error().Err(err).Msg("Failed to encode stats response")
		}
	}
}
