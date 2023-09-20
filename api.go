package main

import (
	"errors"
	"fmt"
	"image"
	"strconv"

	"net/http"

	"github.com/johan-st/go-image-server/units/size"
	"github.com/johan-st/go-image-server/way"
)

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
			srv.respondJson(w, r, http.StatusInternalServerError, resp{
				Message: "Internal Server Error",
			})
		}
		resp := resp{
			Message:      "listing all available image ids",
			AvailableIds: ids,
		}
		l.Debug(resp)
		srv.respondJson(w, r, http.StatusOK, resp)
	}
}

func (srv *server) handleApiImageDelete() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleApiImageDelete")

	type badReqResp struct {
		Error string `json:"error"`
		Got   string `json:"got"`
		Want  string `json:"want"`
	}

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		id_str := way.Param(r.Context(), "id")
		l.Debug("handling delete image request", "id", id_str)

		id, err := strconv.Atoi(id_str)
		if err != nil {
			l.Error("error while parsing id", "id", id_str, "ParseIntError", err)
			srv.respondJson(w, r, http.StatusBadRequest, badReqResp{
				Error: err.Error(),
				Got:   id_str,
				Want:  "int > 0",
			})
			return
		}

		err = srv.ih.Delete(id)
		if err != nil {
			l.Error("error while deleting image", "id", id, "ImageHandlerError", err)
			srv.respondCode(w, r, http.StatusInternalServerError)
			return
		}
		srv.respondCode(w, r, http.StatusOK)
	}
}

// TODO: handle errors and respond with correct status codes
func (srv *server) handleApiImagePost() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleUpload")
	l.With("version", "1")
	l.With("method", "POST")

	type responseOK struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Id      int    `json:"id"`
		Url     string `json:"url"`
	}

	type responseErr struct {
		Status int    `json:"status"`
		Error  string `json:"error"`
	}
	// handler
	// TODO: figure out which erorrs are client errors and which are server errors (warn/info vs error)
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(int64(15 * size.Megabyte))
		if err != nil {
			l.Warn("Error while parsing upload", "ParseMultipartFormError", err)
			srv.respondJson(w, r, http.StatusBadRequest, responseErr{
				Status: http.StatusBadRequest,
				Error:  err.Error(),
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
		headerSize := size.S(header.Size)
		maxUploadSize, err := size.Parse(srv.conf.MaxUploadSize)
		if err != nil {
			l.Fatal("Error while parsing max upload size", "ParseSizeError", err)
			srv.respondJson(w, r, http.StatusInternalServerError, responseErr{
				Status: http.StatusInternalServerError,
				Error:  "Internal Server Error",
			})
			return
		}
		if headerSize > maxUploadSize {
			l.Warn("Maximum upload size exceeded", "FileSize", headerSize, "MaxUploadSize", maxUploadSize)
			srv.respondJson(w, r, http.StatusBadRequest, responseErr{Status: http.StatusBadRequest, Error: "Maximum upload size exceeded"})
			return
		}

		// add to image handler
		id, err := srv.ih.Add(upload)
		if err != nil {
			if errors.Is(err, image.ErrFormat) {
				l.Warn("Error while adding image to handler", "AddIOError", err)
				srv.respondJson(w, r, http.StatusBadRequest, responseErr{
					Status: http.StatusBadRequest,
					Error:  "File is not a valid image",
				})
				return
			}
			l.Error("Error while adding image to handler", "AddIOError", err)
		}

		l.Info("File Uploaded Successfully", "assigned id", id, "original filename", header.Filename, "upload size", header.Size)

		// var url string
		// if srv.conf.Port == 0 || srv.conf.Port == 80 {
		// 	url = fmt.Sprintf("http://%s/%d", srv.conf.Host, id)
		// } else {
		// 	url = fmt.Sprintf("http://%s:%d/%d", srv.conf.Host, srv.conf.Port, id)
		// }

		response := responseOK{
			Status:  http.StatusCreated,
			Message: "File Uploaded Successfully",
			Id:      id,
			Url:     fmt.Sprintf("/%d", id),
		}

		srv.respondJson(w, r, http.StatusCreated, response)
	}
}
