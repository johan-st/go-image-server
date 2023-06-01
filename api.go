package main

import (
	"encoding/json"

	"net/http"

	"github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
	"github.com/johan-st/go-image-server/way"
)

func (srv *server) handleApiGet() http.HandlerFunc {
	// setup
	var l *log.Logger
	if srv.l == nil {
		l = log.Default().With("handler", "handleApiGet")
		l.Warn("no logger provided, using default logger", "level", l.GetLevel()) //DEBUG: remove or move
	} else {
		l = srv.l.With("handler", "handleApiGet")
		l.Warn("logger provided, using provided logger", "level", l.GetLevel()) //DEBUG: remove or move
	}
	l.With("version", "1")
	l.With("method", "GET")

	type resp struct {
		Status   int    `json:"status"`
		Method   string `json:"method"`
		Version  string `json:"version"`
		Message  string `json:"message"`
		Resource string `json:"resource"`
		Id       string `json:"id"`
	}

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		resource := way.Param(r.Context(), "resource")
		id := way.Param(r.Context(), "id")

		resp := resp{
			Status:   http.StatusNotImplemented,
			Method:   "GET",
			Version:  "1",
			Message:  "NOT YET IMPLEMENTED",
			Resource: resource,
			Id:       id,
		}
		l.Debug(resp)
		respondJson(w, r, http.StatusNotImplemented, resp)
	}
}

func (srv *server) handleApiPost() http.HandlerFunc {
	// setup
	var l *log.Logger
	if srv.l == nil {
		l = log.Default().With("handler", "handleApiPost")
	} else {
		l = srv.l.With("handler", "handleApiPost")
	}
	l.With("version", "1")
	l.With("method", "POST")

	// handlerspecific types
	type request struct {
		Action       string `json:"action"`
		ResourceType string `json:"resourceType"`
		Id           string `json:"id"`
	}
	type badReqResp struct {
		Status   int     `json:"status"`
		Message  string  `json:"message"`
		Expectes request `json:"expected body"`
	}

	type response struct {
		Status   int    `json:"status"`
		Method   string `json:"method"`
		Version  string `json:"version"`
		Message  string `json:"message"`
		Resource string `json:"resource"`
		Id       string `json:"id"`
	}

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		resource := way.Param(r.Context(), "resource")
		id := way.Param(r.Context(), "id")

		req := request{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			resp := badReqResp{
				Status:  http.StatusBadRequest,
				Message: "BAD REQUEST. expected json body",
				Expectes: request{
					Action:       "string",
					ResourceType: "image | default | preset | cache",
					Id:           "string",
				},
			}

			l.Info(err)
			respondJson(w, r, http.StatusBadRequest, resp)
			return
		}

		resp := response{
			Status:   http.StatusNotImplemented,
			Method:   "POST",
			Version:  "1",
			Message:  "NOT YET IMPLEMENTED",
			Resource: resource,
			Id:       id,
		}
		l.Debug(resp)
		respondJson(w, r, http.StatusNotImplemented, resp)
	}

}

// TODO: handle errors and respond with correct status codes
func (srv *server) handleUpload(maxSize images.Size) http.HandlerFunc {
	// setup
	var l *log.Logger
	if srv.l == nil {
		l = log.Default().With("handler", "handleUpload")
	} else {
		l = srv.l.With("handler", "handleUpload")
	}
	l.With("version", "1")
	l.With("method", "POST")

	type response struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Id      int    `json:"id,omitempty"`
	}

	// handler
	// TODO: figure out which erorrs are client errors and which are server errors (warn/info vs error)
	return func(w http.ResponseWriter, r *http.Request) {
		// parse up to maxSize
		err := r.ParseMultipartForm(int64(maxSize))
		if err != nil {
			l.Warn("Error while parsing upload", "ParseMultipartFormError", err)
			return
		}

		// Get header for filename, size and headers
		upload, header, err := r.FormFile("image")
		if err != nil {
			l.Warn("Error Retrieving the File", "FormFileError", err)
			return
		}
		defer upload.Close()

		// add to image handler
		id, err := srv.ih.Add(upload)
		if err != nil {
			l.Error("Error while adding image to handler", "AddIOError", err)
		}

		l.Info("File Uploaded Successfully", "assigned id", id, "original filename", header.Filename, "upload size", header.Size)
		response := response{
			Status:  http.StatusCreated,
			Message: "File Uploaded Successfully",
			Id:      id,
		}

		respondJson(w, r, http.StatusCreated, response)
	}
}

// HELPER
func respondJson(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
