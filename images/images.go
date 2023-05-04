package images

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/nfnt/resize"
	//github.com/hashicorp/golang-lru/v2
	// TODO: consider using hashicorp lru or ARC for cache
	// build my own cache will be a good exercise. If I match the interface of hashicorp lru,
	// I can easily switch to it later. Or even let the user of this package give me a lru
	// eqvivalent on creation.
)

// TODO: create interface of used mathods to facilitate injecting a logger
// type Logger interface {
// 	Debug(msg interface{}, keyval ...interface{})
// 	Debugf(format string, args ...interface{})
// 	Info(msg interface{}, keyval ...interface{})
// 	Infof(format string, args ...interface{})
// 	Warn(msg interface{}, keyval ...interface{})
// 	Warnf(format string, args ...interface{})
// 	Error(msg interface{}, keyval ...interface{})
// 	Errorf(format string, args ...interface{})
// 	Fatal(msg interface{}, keyval ...interface{})
// 	Fatalf(format string, args ...interface{})
//  With
// }

const (
	DefaultOriginalsDir = "img/originals"
	DefaultCacheDir     = "img/cache"
	commonExt           = ".jpg" //somewhat of a hack. all files are saved as '*.jpg' TODO: clean up
)

// ImageHandler is the main type of this package.
type ImageHandler struct {
	conf     Config  // TODO: Read-only?
	latestId ImageId // TODO: this can be accessed by multiple goroutines. Make thread-safe.
	l        *log.Logger
}

// Config represents the configuration of an ImageHandler.
// unset (0/"") parameters will be considered as "use default".
type Config struct {
	OriginalsDir  string          // path to originals			(default: "img/originals")
	CacheDir      string          // path to cache			(default: "img/cache")
	DefaultParams ImageParameters // default image parameters		(default: see ImageParameters)
	CacheRules    CacheRules      // cache rules			(default: see CacheRules)
	CreateDirs    bool            // create directories if needed	(default: false)
	SetPerms      bool            // set permissions if needed		(default: false)
}

// ImageParameters represents how an image should be pressented.
// note: Use 0 (zero) to explicitly set default.
type ImageParameters struct {
	Format  Format // Jpeg, Png, Gif		(default: Jpeg)
	Width   uint   // width in pixels 		(default: match original)
	Height  uint   // height in pixels		(default: match original)
	Quality int    // Jpeg:1-100, Gif:1-256	(default: jpeg: 80, gif: 256)
	MaxSize Size   // Max file size in bytes	(default: Infinite)
}

func (ip *ImageParameters) String() string {
	return fmt.Sprintf("%dx%d_q%d_%d", ip.Width, ip.Height, ip.Quality, ip.MaxSize)
}

// ImageHandler will try to keep the cache within these limits but does not guarantee it.
//
// note: Use 0 (zero) to explicitly set default.
type CacheRules struct {
	MaxTimeSinceUse time.Duration // Max age in seconds		(default: 0, unlimited)
	MaxSize         Size          // Max cache size in bytes	(default: 1 Gigabyte)
}

type ImageId int

func (id ImageId) String() string {
	return strconv.Itoa(int(id))
}

// Format represents image formats.
type Format int

const (
	Jpeg Format = iota // quality 1-100
	Png                // always lossless
	Gif                // num colors 1-256
)

// Represents file sizes:
//   - Infinite = 0
//   - 1 Kilobyte = 1024 bytes
//   - 1 Megabyte = 1024 Kilobytes
//   - 1 Gigabyte = 1024 Megabytes
//   - 1 Terabyte = 1024 Gigabytes
//   - 1 Petabyte = 1024 Terabytes
type Size uint64

const (
	Infinite  Size = 0                // no limit
	Kilobytes      = 1024             // 1 Kilobyte = 1024 bytes
	Megabytes      = 1024 * Kilobytes // 1 Megabyte = 1024 Kilobytes
	Gigabytes      = 1024 * Megabytes // 1 Gigabyte = 1024 Megabytes
	Terabytes      = 1024 * Gigabytes // 1 Terabyte = 1024 Gigabytes
	Petabytes      = 1024 * Terabytes // 1 Petabyte = 1024 Terabytes
)

type ErrIdNotFound struct {
	IdGiven ImageId
	Err     error
}

func (e ErrIdNotFound) Error() string {
	return fmt.Sprintf("id (%d) not found\nerror: %s", e.IdGiven, e.Err.Error())
}

func (e ErrIdNotFound) Is(err error) bool {
	_, ok := err.(ErrIdNotFound)
	return ok
}

// New creates a new Imageandler with the given configuration.
// Caller is responsible for running CacheHousekeeping() periodically to trigger cache cleanup.
// Preferably when the server is not under heavy load.
func New(conf Config, l *log.Logger) (*ImageHandler, error) {
	err := checkDirs(conf)
	if err != nil {
		return &ImageHandler{}, err
	}
	if l == nil {
		l = log.New(os.Stderr).WithPrefix("Image Handler")
		l.SetLevel(logLevel())
		l.Info("Using default logger")
	}

	l.Debug("New", "Config", conf)
	return &ImageHandler{
			conf:     conf,
			latestId: findLatestId(conf.OriginalsDir),
			l:        l},
		nil
}

// returns the path to the processed image.
func (h *ImageHandler) Get(params ImageParameters, id ImageId) (string, error) {
	cachePath := h.cachePath(params, id)
	h.l.Info("Get", "ImageId", id, "ImageParameters", params, "cachePath", cachePath)

	_, err := os.Stat(cachePath)
	// file exists
	if err == nil {
		return cachePath, nil
	}

	// error other than file does not exist
	if !os.IsNotExist(err) {
		return "", err
	}

	// file does not exist
	err = h.processAndCache(params, id, cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrIdNotFound{IdGiven: id, Err: err}
		}
		return "", err
	}

	// file was created
	return cachePath, nil
}

func (h *ImageHandler) Add(path string) (ImageId, error) {
	h.l.Info("Add", "path", path)

	// check if file exists
	srcf, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer srcf.Close()

	// check if file is a supported image
	ext := filepath.Ext(path)
	supext := []string{".jpg", ".jpeg", ".png", ".gif"}
	if !contains(supext, ext) {
		return 0, fmt.Errorf("unsupported image format")
	}
	// TODO: check with the image package instead of just the extension?
	// _, _, err = image.Decode(srcf)
	// if err != nil {
	// 	return 0, err
	// }

	// get next id
	nextId := h.latestId + 1
	h.latestId = nextId

	// determine destination path (TODO: should extention be changed?)
	dst := h.conf.OriginalsDir + "/" + nextId.String() + ".jpg"

	// copy file to originals
	dstf, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, err
	}
	defer dstf.Close()
	_, err = io.Copy(dstf, srcf)
	if err != nil {
		return 0, err
	}
	// return id
	return nextId, nil
}

func (h *ImageHandler) Remove(id ImageId) error {
	h.l.Info("Remove", "id", id)
	// remove from cache
	h.CacheClearFor(id)
	// remove original
	oPath := h.originalPath(id)
	err := os.Remove(oPath)
	if err != nil {
		return err
	}

	return nil
}

//  TODO: decide which of these should even exist

// clear all cache
func (h *ImageHandler) CacheClear() (int, error) {
	h.l.Debug("CacheClear")
	// cachefoldersize
	dir, err := os.Open(h.conf.CacheDir)
	if err != nil {
		h.l.Error("CacheClear", "error", err)
		return 0, err
	}
	defer dir.Close()

	// get list of files
	files, err := dir.Readdir(0)
	if err != nil {
		h.l.Error("CacheClear", "error", err)
		return 0, err
	}

	totalBytes := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		totalBytes += int(f.Size())
		// remove files
		err = os.Remove(h.conf.CacheDir + "/" + f.Name())
		if err != nil {
			h.l.Error("CacheClear", "error", err)
			return 0, err
		}
	}
	h.l.Info("Cached cleared", "freed Bytes", totalBytes)
	return totalBytes, nil
}

func (h *ImageHandler) CacheClearFor(id ImageId) (int, error) {
	h.l.Debug("id", id)

	bytesFreed := 0
	dir, err := os.Open(h.conf.CacheDir)
	if err != nil {
		h.l.Error("CacheClearFor", "error", err)
		return 0, err
	}
	defer dir.Close()

	// get list of files
	files, err := dir.Readdir(0)
	if err != nil {
		h.l.Error("CacheClearFor", "error", err)
		return 0, err
	}

	errs := []error{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// remove files
		if strings.HasPrefix(f.Name(), id.String()) {
			bytesFreed += int(f.Size())
			err = os.Remove(h.conf.CacheDir + "/" + f.Name())
			if err != nil {
				h.l.Error("CacheClearFor", "error", err)
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return bytesFreed, fmt.Errorf("errors while removing files. errors: %s", errs)
	}

	return bytesFreed, nil
}

// TODO: page and chunk as arguments
func (h *ImageHandler) ListIds() ([]ImageId, error) {
	h.l.Debug("ListIds")

	dir, err := os.Open(h.conf.OriginalsDir)
	if err != nil {
		h.l.Error("ListIds", "error", err)
		return nil, err
	}
	defer dir.Close()

	// get list of files
	files, err := dir.Readdir(0)
	if err != nil {
		h.l.Error("ListIds", "error", err)
		return nil, err
	}

	ids := []ImageId{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// add id to list
		idStr := strings.Split(f.Name(), ".")[0]
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			h.l.Error("ListIds", "error", err)
			return nil, err
		}
		id := ImageId(idInt)

		h.l.Debug("ListIds got", "id", id, "from", f.Name()) //DEBUG: remove this
		ids = append(ids, id)
	}
	return ids, nil
}

// TODO: should probably lock the cache while doing this
func (h *ImageHandler) CacheHouseKeeping() (int, error) {

	// sort by last access time
	// List files to remove
	// lock cache
	// remove files
	// unlock cache

	h.l.Error("CacheHouseKeeping is not yet implementes")
	return 0, fmt.Errorf("not implemented")
}

func findLatestId(originalsPath string) ImageId {
	return 0
}

// Create a new image with the given configuration and store it in the cache.
// return the path to the cached image.
// TODO: quality 0 should be default
func (h *ImageHandler) processAndCache(params ImageParameters, id ImageId, cachePath string) error {
	oPath := h.originalPath(id)
	oImg, err := loadImage(oPath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	img := resize.Resize(params.Width, params.Height, oImg, resize.Lanczos3)

	switch params.Format {
	case Jpeg:
		if params.Quality == 0 {
			params.Quality = 90
		}
		opt := &jpeg.Options{Quality: params.Quality}
		err = jpeg.Encode(file, img, opt)
		if err != nil {
			return err
		}
	case Png:
		err = png.Encode(file, img)
		if err != nil {
			return err
		}
	case Gif:
		if params.Quality == 0 {
			params.Quality = 256
		}
		nc := params.Quality
		opt := &gif.Options{NumColors: nc}
		err = gif.Encode(file, img, opt)
		if err != nil {
			return err
		}
	}
	return nil
}
func (h *ImageHandler) originalPath(id ImageId) string {
	return filepath.Join(h.conf.OriginalsDir, id.String()+commonExt)
}

func (h *ImageHandler) cachePath(params ImageParameters, id ImageId) string {
	return filepath.Join(h.conf.CacheDir, id.String()+"_"+params.String()+commonExt)
}

// HELPER

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// NOTE: is duplicated in main
func logLevel() log.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "WARN":
		return log.WarnLevel
	case "ERROR":
		return log.ErrorLevel
	case "FATAL":
		return log.FatalLevel
	}
	return log.ErrorLevel
}

// retunes the image specified by the path
func loadImage(path string) (image.Image, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func checkDirs(c Config) error {
	paths := []string{c.OriginalsDir, c.CacheDir}
	err := checkExists(paths, c.CreateDirs)
	if err != nil {
		return err
	}
	return checkFilePermissions(paths, c.SetPerms)
}

func checkExists(paths []string, createDirs bool) error {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if createDirs {
				err = os.MkdirAll(path, 0700)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("directory %s does not exist", path)
			}
		}
	}
	return nil
}

// check that the images directory exists and is writable. If not, set up needed permissions.
// TODO: folders should be configurable
// TODO: check perm bool
func checkFilePermissions(paths []string, setPerms bool) error {
	for _, path := range paths {
		err := filepath.WalkDir(path, permAtLeast(0700, 0600))
		if err != nil {
			return err
		}
	}
	return nil
}

// Will extend permissions if needed, but will not reduce them.
func permAtLeast(dir os.FileMode, file os.FileMode) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i, err := d.Info()
		if err != nil {
			return err
		}
		perm := os.FileMode(0)
		if d.IsDir() {
			perm = dir
		} else {
			perm = file
		}

		if i.Mode().Perm()&perm < perm {
			p := i.Mode().Perm() | perm
			err := os.Chmod(path, p)
			if err != nil {
				return fmt.Errorf("'%s' has insufficient permissions and setting new permission failed. (was: %o, need at least: %o)", path, i.Mode().Perm(), perm)
			}
			fmt.Printf("'%s' had insufficient permissions. Setting permissions to %o. (was: %o, need at least: %o)\n", path, p, i.Mode().Perm(), perm)

		}
		return nil
	}
}
