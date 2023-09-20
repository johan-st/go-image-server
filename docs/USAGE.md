# `jst_ImageServer`

a go image server

## NOTICE!

This server is still under development. Incomplete features and issues are to be expected.

## routes

### GET /

This documentation resides here

### GET /:image_id

Requests to any path beyond root ("/") are treated as an image request.
This path returns the original image by id

### GET /:image_id/:desired_filename.jpeg

Subsequent path does not change the response but is helpfull for naming the file fetched.
The titular example return a file named desired_filname.jpg

## Preprocessing

### How to get the size you want?

Describing the image you want is done through query parameters added to the url.

#### parameter

| parameter       | type    | range                         | interpretation                                  |
| --------------- | ------- | ----------------------------- | ----------------------------------------------- |
| `w` / `width`   | integer | 1 or greater                  | desired width in pixels                         |
| `h` / `height`  | integer | 1 or greater                  | desired height in pixels                        |
| `f` / `format`  | string  | "jpeg" / "jpg", "png","gif"   | desired image format                            |
| `q` / `quality` | integer | 1-100 for jpeg. 1-256 for gif | jpeg: quality in percent. gif: number of colors |

#### parameters details:
- `width` / `w`: Accepts integers greater than 0. This parameter determines the width in pixels of the returned image. 
- `height` / `h`: Accepts integers greater than 0. This parameter determines the height in pixels of the returned image. 
  - If only one of width or height is specified the other will be calculated to keep the aspect ratio of the original image.
  - If both are specified the image will be cropped to the specified size. (TODO: make it crop, not stretch)
- `format` / `f`: Accepts "jpeg"/"jpg", "png" and "gif". This parameter determines the format of the returned image. 
- `quality` / `q`: quality, accepts integers. 
  - `Jpeg`: Accepts values between 1 and 100 (inclusive). Around 80 is a good value for most images.
  - `png`: Can not be compressed and will always be full quality (TODO: source)
  - `gif`: Quality is determined by the number of colors in the image. Accepts values between 1 and 256 (inclusive).



## examples

### /1/linked_image.jpeg?w=100
![linked image example](/1/linked_image.jpeg?w=200)


### /2/linked_image.jpeg?f=png&w=25&h=25
![linked image example](/2/linked_image.jpeg?f=png&w=250&h=250)

### /3?f=gif&q=4&w=700&h=100
![linked image example](/3?f=gif&q=4&w=700&h=100)


## example code 

```go
package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	log "github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
	components "github.com/johan-st/go-image-server/pages/components"
	"github.com/johan-st/go-image-server/units/size"
	"github.com/johan-st/go-image-server/way"
)

//go:embed pages/assets
var staticFS embed.FS
var (
	darkTheme = components.Theme{
		ColorPrimary:    "#f90",
		ColorSecondary:  "#fa3",
		ColorBackground: "#333",
		ColorText:       "#aaa",
		ColorBorder:     "#666",
		BorderRadius:    "1rem",
	}
	
	metadata = map[string]string{
		"Description": "img.jst.dev is a way for Johan Strand to learn more Go and web development.",
		"Keywords":    "image, hosting",
		"Author":      "Johan Strand",
	}
)


type server struct {
	errorLogger  *log.Logger // *required
	accessLogger *log.Logger // optional

	conf   confHttp
	ih     *images.ImageHandler
	router way.Router

	// TODO: make concurrent safe
	Stats struct {
		StartTime    time.Time
		Requests     int
		Errors       int
		ImagesServed int
	}
}

// Register handlers for routes
func (srv *server) routes() {

	srv.Stats.StartTime = time.Now()

	// Docs / root
	if srv.conf.Docs {
		srv.router.HandleFunc("GET", "", srv.handleDocs())
	}

	// STATIC ASSETS
	srv.router.HandleFunc("GET", "/favicon.ico", srv.handleFavicon())
	srv.router.HandleFunc("GET", "/assets/", srv.handleAssets())

	// API
	srv.router.HandleFunc("GET", "/api/images", srv.handleApiImageGet())
	srv.router.HandleFunc("POST", "/api/images", srv.handleApiImagePost())
	srv.router.HandleFunc("DELETE", "/api/images/:id", srv.handleApiImageDelete())
	srv.router.HandleFunc("*", "/api/", srv.handleNotAllowed())

	// Admin
	srv.router.HandleFunc("GET", "/admin", srv.handleAdminTempl())
	srv.router.HandleFunc("GET", "/admin/:page", srv.handleAdminTempl())
	srv.router.HandleFunc("GET", "/admin/images/:id", srv.handleAdminImage())

	// Serve Images
	srv.router.HandleFunc("GET", "/:id/:preset/", srv.handleImgWithPreset())
	srv.router.HandleFunc("GET", "/:id/", srv.handleImg())

	// 404
	srv.router.NotFound = srv.handleNotFound()
}
```