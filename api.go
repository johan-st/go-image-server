package main

import (
	"errors"
	"image"

	"net/http"

	"github.com/johan-st/go-image-server/images"
	"github.com/johan-st/go-image-server/way"
)

// DOCS

func (src *server) handleApiDocs() http.HandlerFunc {
	// setup
	l := src.errorLogger.With("handler", "handlerApiDocs")

	type resp struct {
		Message string `json:"message"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("TODO: implement handlerApiDocs")
		respondJson(w, r, http.StatusNotImplemented, resp{
			Message: "TODO: implement handlerApiDocs",
		})
	}
}

// IMAGES

func (srv *server) handleApiImageGet() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleApiImageGet")

	type resp struct {
		Message      string `json:"message"`
		AvailableIds []int  `json:"availableIds,omitempty"`
	}

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		ids, err := srv.ih.Ids()
		if err != nil {
			l.Error(err)
			respondJson(w, r, http.StatusInternalServerError, resp{
				Message: "Internal Server Error",
			})
		}
		resp := resp{
			Message:      "listing all available image ids",
			AvailableIds: ids,
		}
		l.Debug(resp)
		respondJson(w, r, http.StatusOK, resp)
	}
}

func (srv *server) handleApiImageDelete() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleApiImageDelete")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		id := way.Param(r.Context(), "id")
		l.Debug("handling delete image request", "id", id)

		// err = srv.ih.Delete(req.Id)
		// if err != nil {
		// l.Error("error while deleting image", "id", req.Id, "ImageHandlerError", err)
		// 	respondCode(w, r, http.StatusInternalServerError)
		// 	return
		// }
		// respondCode(w, r, http.StatusOK)
		respondCode(w, r, http.StatusNotImplemented)
	}
}

// TODO: handle errors and respond with correct status codes
func (srv *server) handleUpload() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleUpload")
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
		err := r.ParseMultipartForm(int64(15 * images.Megabyte))
		if err != nil {
			l.Warn("Error while parsing upload", "ParseMultipartFormError", err)
			respondJson(w, r, http.StatusBadRequest, response{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		// Get header for filename, size and headers
		// TODO: can I get the first file from the form without knowing the name?
		upload, header, err := r.FormFile("image")
		if err != nil {
			l.Warn("Error Retrieving the File", "FormFileError", err)
			return
		}
		defer upload.Close()

		// check size
		hs := images.Size(header.Size)
		ms, err := images.ParseSize(srv.conf.MaxUploadSize)
		if err != nil {
			l.Fatal("Error while parsing max upload size", "ParseSizeError", err)
			respondJson(w, r, http.StatusInternalServerError, response{
				Status:  http.StatusInternalServerError,
				Message: "Internal Server Error",
			})
			return
		}
		if hs > ms {
			l.Warn("Maximum upload size exceeded", "FileSize", hs, "MaxUploadSize", ms)
			respondJson(w, r, http.StatusBadRequest, response{Status: http.StatusBadRequest, Message: "Maximum upload size exceeded"})
			return
		}

		// add to image handler
		id, err := srv.ih.Add(upload)
		if err != nil {
			if errors.Is(err, image.ErrFormat) {
				l.Warn("Error while adding image to handler", "AddIOError", err)
				respondJson(w, r, http.StatusBadRequest, response{
					Status:  http.StatusBadRequest,
					Message: "File is not a valid image",
				})
				return
			}
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
