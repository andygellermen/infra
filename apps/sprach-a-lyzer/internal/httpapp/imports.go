package httpapp

import (
	"errors"
	"net/http"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/managedimport"
)

func (a *App) prepareImport(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[managedimport.PrepareRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.imports.Prepare(request.Context(), input)
	if err != nil {
		a.writeImportError(response, err)
		return
	}
	writeVersionedJSON(response, "managed-import-plan", result.ContractVersion, result)
}
func (a *App) commitImport(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[managedimport.CommitRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.imports.Commit(request.Context(), input)
	if err != nil {
		a.writeImportError(response, err)
		return
	}
	writeVersionedJSON(response, "managed-import-operation", result.ContractVersion, result)
}
func (a *App) rollbackImport(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[managedimport.CommitRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.imports.Rollback(request.Context(), input)
	if err != nil {
		a.writeImportError(response, err)
		return
	}
	writeVersionedJSON(response, "managed-import-operation", result.ContractVersion, result)
}
func (a *App) importHistory(response http.ResponseWriter, request *http.Request) {
	result, err := a.imports.History(request.Context())
	if err != nil {
		a.writeImportError(response, err)
		return
	}
	writeVersionedJSON(response, "managed-import-history", managedimport.ContractVersion, result)
}
func (a *App) importAudit(response http.ResponseWriter, request *http.Request) {
	result, err := a.imports.Audit(request.Context(), request.PathValue("batch_id"))
	if err != nil {
		a.writeImportError(response, err)
		return
	}
	writeVersionedJSON(response, "managed-import-audit", managedimport.ContractVersion, result)
}
func (a *App) writeImportError(response http.ResponseWriter, err error) {
	if errors.Is(err, managedimport.ErrForbidden) {
		writeError(response, http.StatusForbidden, "IMPORT_FORBIDDEN", err.Error())
		return
	}
	if errors.Is(err, managedimport.ErrNotReady) || errors.Is(err, managedimport.ErrValidateOnly) || errors.Is(err, managedimport.ErrAlreadyRolledBack) {
		writeError(response, http.StatusConflict, "IMPORT_STATE_CONFLICT", err.Error())
		return
	}
	writeError(response, http.StatusUnprocessableEntity, "MANAGED_IMPORT_FAILED", err.Error())
}
