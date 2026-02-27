package shorter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	domainmodel "github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func APIHandler(usecase usecase.LinkCreateUsecase, auditor service.AuditorService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get(common.ContentType) != common.ApplicationJSON {
			common.MethodNotAllowedError(writer)
			return
		}

		var shortenRequest model.ShortenRequest
		dec := json.NewDecoder(request.Body)
		if err := dec.Decode(&shortenRequest); err != nil {
			common.BadRequestError(writer, err)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(request.Body)

		auditURL := auditURLFromRaw(shortenRequest.URL)
		userID := authctx.UserID(request.Context())

		shortKey, err := GenerateShortKey(shortenRequest.URL, request.Context(), writer, usecase)
		if err != nil {
			var conflictErr *domainmodel.URLConflictError
			if errors.As(err, &conflictErr) {
				var response = &model.ShortenResponse{
					Result: GenerateShortenURL(usecase.GetBaseURL(), conflictErr.ExistingShortKey, request),
				}
				bytes, jsonErr := response.MarshalJSON()
				if jsonErr != nil {
					common.InternalError(writer, jsonErr)
					return
				}

				writer.Header().Set(common.ContentType, common.ApplicationJSON)
				writer.WriteHeader(http.StatusConflict)
				_, writeError := writer.Write(bytes)
				if writeError != nil {
					common.InternalError(writer, writeError)
				} else if auditor != nil {
					auditor.AuditNewEvent(request.Context(), domainmodel.AuditEvent{
						Timestamp: time.Now().Unix(),
						Action:    "shorten",
						UserID:    userID,
						URL:       auditURL,
					})
				}
				return
			}
			return
		}

		var response = &model.ShortenResponse{
			Result: GenerateShortenURL(usecase.GetBaseURL(), shortKey, request),
		}
		bytes, jsonErr := response.MarshalJSON()
		if jsonErr != nil {
			common.InternalError(writer, jsonErr)
			return
		}

		writer.Header().Set(common.ContentType, common.ApplicationJSON)
		writer.WriteHeader(http.StatusCreated)
		_, writeError := writer.Write(bytes)
		if writeError != nil {
			common.InternalError(writer, writeError)
			return
		}
		if auditor != nil {
			auditor.AuditNewEvent(request.Context(), domainmodel.AuditEvent{
				Timestamp: time.Now().Unix(),
				Action:    "shorten",
				UserID:    userID,
				URL:       auditURL,
			})
		}
	}
}
