package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	zlog "github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/schema-repository/internal/service"
)

// Handler struct holds dependencies, like the schema service.
type Handler struct {
	SchemaSvc *service.SchemaService
}

// NewHandler creates a new Handler instance.
func NewHandler(svc *service.SchemaService) *Handler {
	return &Handler{SchemaSvc: svc}
}

// CreateSchemaHandler handles the creation of a new schema.
func (h *Handler) CreateSchemaHandler(ctx *gin.Context) {
	var reqURI SchemaRequestURI
	if err := ctx.ShouldBindUri(&reqURI); err != nil {
		zlog.Warn().Msg("failed to bind request URI")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var req SchemaBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Warn().Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.SchemaSvc.AddSchema(ctx, reqURI.Name, reqURI.Version, req.Schema); err != nil {
		zlog.Err(err).Msg("internal server error")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "An unexpected error occurred while persisting the schema"})
		return
	}

	ctx.Status(http.StatusCreated)
}

// GetSchemaHandler handles the retrieval of a schema.
func (h *Handler) GetSchemaHandler(ctx *gin.Context) {
	var reqURI SchemaRequestURI
	if err := ctx.ShouldBindUri(&reqURI); err != nil {
		zlog.Warn().Msg("failed to bind request URI")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	schema, err := h.SchemaSvc.GetSchema(ctx, reqURI.Name, reqURI.Version)
	if err != nil {
		errStr := err.Error()
		if errStr == service.ErrorSchemaNotFound {
			zlog.Warn().Str("schema", reqURI.Name).Str("version", reqURI.Version.String()).Msg("schema not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{Error: errStr})
		} else {
			zlog.Err(err).Msg("internal server error")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "An unexpected error occurred while retrieving the schema"})
		}
		return
	}

	ctx.JSON(http.StatusOK, SchemaBody{
		Schema: schema,
	})
}

// DeleteSchemaHandler handles the retrieval of a schema.
func (h *Handler) DeleteSchemaHandler(ctx *gin.Context) {
	var reqURI SchemaRequestURI
	if err := ctx.ShouldBindUri(&reqURI); err != nil {
		zlog.Warn().Msg("failed to bind request URI")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.SchemaSvc.DeleteSchema(ctx, reqURI.Name, reqURI.Version)
	if err != nil {
		errStr := err.Error()
		if errStr == service.ErrorSchemaNotFound {
			zlog.Warn().Str("schema", reqURI.Name).Str("version", reqURI.Version.String()).Msg("schema not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{Error: errStr})
		} else {
			zlog.Err(err).Msg("internal server error")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "An unexpected error occurred while deleting the schema"})
		}
		return
	}

	ctx.Status(http.StatusOK)
}
