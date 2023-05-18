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

// cache is expected to be thread-safe. It should return true if the path is in cache and false otherwise.
//
// It should add the path to cache if it is not already there.
//
// Cach should send evicted paths to the image handler to be deleted from disk.
type cache interface {
	Access(path string) bool
}

const (
	originalsExt = ".jpg" //somewhat of a hack. all originals are saved and retrieved with this extention TODO: find a better way?
)

// ImageHandler is the main type of this package.
type ImageHandler struct {
	conf     Config // TODO: Read-only?
	latestId int    // TODO: this can be accessed by multiple goroutines. Make thread-safe.
	l        *log.Logger
	cache    cache //TODO: slice as initial prototype
	// TODO: might need a map to keep track of ids are in use after implementing "delete"
}

// Config represents the configuration of an ImageHandler.
// unset (0/"") parameters will be considered as "use default".
//
// Default values are:
//
//	OriginalsDir: "img/originals"
//	CacheDir: "img/cache"
//	DefaultParams: see ImageParameters
//	CreateDirs: false
//	SetPerms: false
type Config struct {
	OriginalsDir  string          // path to originals
	CacheDir      string          // path to cache
	DefaultParams ImageParameters // default image parameters
	CreateDirs    bool            // create directories if needed
	SetPerms      bool            // set permissions if needed
}

func (c *Config) withDefault() {
	if c.OriginalsDir == "" {
		c.OriginalsDir = "img/originals"
	}
	if c.CacheDir == "" {
		c.CacheDir = "img/cache"
	}
	if c.DefaultParams.Format == "" {
		c.DefaultParams.Format = Jpeg
	}
	if c.DefaultParams.Width == 0 && c.DefaultParams.Height == 0 {
		c.DefaultParams.Width = 800
	}
	if c.DefaultParams.Quality == 0 && c.DefaultParams.Format == Jpeg {
		c.DefaultParams.Quality = 80
	}
	if c.DefaultParams.Quality == 0 && c.DefaultParams.Format == Gif {
		c.DefaultParams.Quality = 256
	}
}

// ImageParameters represents how an image should be pressented.
// note: Use 0 (zero) to explicitly set default.
//
// Default values are set when ImageHandler is created.
//
// Default values are:
//
//	Format: Jpeg
//	Width: 800
//	Height: 0 (keep aspect ratio)
//	Quality: 80 (Jpeg), 256 (Gif)
//	MaxSize: 0 (infinite)
type ImageParameters struct {
	Format  Format // Jpeg, Png, Gif (default: Jpeg)
	Width   uint   // width in pixels (default: 800)
	Height  uint   // height in pixels (default: 0/keep aspect ratio)
	Quality int    // Jpeg:1-100, Gif:1-256 (default: 80 for jpeh, 256 for gif)
	MaxSize Size   // Max file size in bytes (default: 0/infinite)
}

func (ip *ImageParameters) String() string {
	if ip.Format == "" {
		ip.Format = Jpeg
	}
	return fmt.Sprintf("%dx%d_q%d_%d.%s", ip.Width, ip.Height, ip.Quality, ip.MaxSize, ip.Format)
}

func (ip *ImageParameters) withDefaults(def ImageParameters) ImageParameters {
	if ip.Format == "" {
		ip.Format = def.Format
	}
	if ip.Format == Jpeg && ip.Quality == 0 {
		ip.Quality = def.Quality
	}
	if ip.Format == Gif && ip.Quality == 0 {
		ip.Quality = 256
	}
	if ip.Width == 0 {
		ip.Width = def.Width
	}
	if ip.Height == 0 {
		ip.Height = def.Height
	}
	if ip.MaxSize == 0 {
		ip.MaxSize = def.MaxSize
	}
	return *ip
}

// Format represents image formats.
type Format string

const (
	Jpeg Format = "jpeg" // quality 1-100
	Png  Format = "png"  // always lossless
	Gif  Format = "gif"  // num colors 1-256
)

func (f Format) String() string {
	return string(f)
}

type ErrIdNotFound struct {
	IdGiven int
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
//
// TODO: should take a list of parameters to create for all images
// DEBUG: TODO: MUST create a cache based on files in cache folder on startup
func New(conf Config, l *log.Logger) (*ImageHandler, error) {
	conf.withDefault()

	err := checkDirs(conf)
	if err != nil {
		return &ImageHandler{}, err
	}
	if l == nil {
		l = log.New(os.Stderr).WithPrefix("Image Handler")
		l.SetLevel(logLevel())
		l.Info("Using default logger")
	}

	evitedChan := make(chan string, 10)
	go fileRemover(l, evitedChan)

	l.Debug("Creating new ImageHandler", "Config", conf)
	ih := ImageHandler{
		conf:     conf,
		latestId: 0,
		l:        l,
		cache:    NewLru(10, evitedChan),
	}

	ih.latestId, err = ih.findLatestId()
	if err != nil {
		ih.l.Fatal("could not get latest id during setup.", "error", err)
		return nil, err
	}
	return &ih, nil
}

// returns the path to the processed image.
func (h *ImageHandler) Get(params ImageParameters, id int) (string, error) {
	// normalize parameters with defaults
	params = params.withDefaults(h.conf.DefaultParams)

	cachePath := h.cachePath(params, id)
	h.l.Info("Get", "ImageId", id, "ImageParameters", params, "cachePath", cachePath)

	// Look for the image in the cache, return it if it does
	if inCache := h.cache.Access(cachePath); inCache {
		return cachePath, nil
	}

	// file does not exist
	size, err := h.createImage(params, id, cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrIdNotFound{IdGiven: id, Err: err}
		}
		return "", err
	}

	// file was created
	h.l.Debug("Cachefile Created", "path", cachePath, "size", size)
	return cachePath, nil
}

func (h *ImageHandler) Add(path string) (int, error) {
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

	// update latest id
	h.latestId++

	// determine destination path (TODO: should extention be changed?)
	dst := h.conf.OriginalsDir + "/" + strconv.Itoa(h.latestId) + originalsExt

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
	return h.latestId, nil
}

// func (h *ImageHandler) Remove(id ImageId) error {
// 	h.l.Info("Remove", "id", id)
// 	// remove from cache
// 	h.CacheClearFor(id)
// 	// remove original
// 	oPath := h.originalPath(id)
// 	err := os.Remove(oPath)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

//  TODO: decide which of these should even exist

// // clear all cache
// func (h *ImageHandler) CacheClear() (Size, error) {
// 	h.l.Debug("CacheClear")
// 	h.cache.clear()
// 	// cachefoldersize
// 	dir, err := os.Open(h.conf.CacheDir)
// 	if err != nil {
// 		h.l.Error("CacheClear", "error", err)
// 		return 0, err
// 	}
// 	defer dir.Close()

// 	// get list of files
// 	files, err := dir.Readdir(0)
// 	if err != nil {
// 		h.l.Error("CacheClear", "error", err)
// 		return 0, err
// 	}

// 	totalBytes := 0
// 	for _, f := range files {
// 		if f.IsDir() {
// 			continue
// 		}
// 		totalBytes += int(f.Size())
// 		// remove files
// 		err = os.Remove(h.conf.CacheDir + "/" + f.Name())
// 		if err != nil {
// 			h.l.Error("CacheClear", "error", err)
// 			return 0, err
// 		}
// 	}
// 	h.l.Info("Cached cleared", "freed Bytes", totalBytes)
// 	return Size(totalBytes), nil
// }

// func (h *ImageHandler) CacheClearFor(id ImageId) (Size, error) {
// 	h.l.Debug("id", id)

// 	bytesFreed := 0
// 	dir, err := os.Open(h.conf.CacheDir)
// 	if err != nil {
// 		h.l.Error("CacheClearFor", "error", err)
// 		return 0, err
// 	}
// 	defer dir.Close()

// 	// get list of files
// 	files, err := dir.Readdir(0)
// 	if err != nil {
// 		h.l.Error("CacheClearFor", "error", err)
// 		return 0, err
// 	}

// 	errs := []error{}
// 	for _, f := range files {
// 		if f.IsDir() {
// 			continue
// 		}
// 		// remove files
// 		if strings.HasPrefix(f.Name(), id.String()) {
// 			bytesFreed += int(f.Size())
// 			err = os.Remove(h.conf.CacheDir + "/" + f.Name())
// 			if err != nil {
// 				h.l.Error("CacheClearFor", "error", err)
// 				errs = append(errs, err)
// 			}
// 		}
// 	}
// 	if len(errs) > 0 {
// 		return Size(bytesFreed), fmt.Errorf("errors while removing files. errors: %s", errs)
// 	}

// 	return Size(bytesFreed), nil
// }

// TODO: page and chunk as arguments for when we have thousands of ids?
func (h *ImageHandler) ListIds() ([]int, error) {
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

	ids := []int{}
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
		id := idInt

		h.l.Debug("ListIds got", "id", id, "from", f.Name())
		ids = append(ids, id)
	}
	return ids, nil
}

// TODO: should probably lock the cache while doing this
// func (h *ImageHandler) CacheHouseKeeping() (Size, error) {

// 	paths := h.cache.cacheByRules(h.conf.CacheRules)
// 	bytesFreed := int64(0)
// 	var errs []error
// 	for _, p := range paths {
// 		h.l.Debug("CacheHouseKeeping", "removing", p)
// 		fileStat, err := os.Stat(p)
// 		if err != nil {
// 			h.l.Error("CacheHouseKeeping", "error", err)
// 			errs = append(errs, err)
// 			continue
// 		}
// 		size := fileStat.Size()
// 		if err != nil {
// 			h.l.Error("CacheHouseKeeping", "error", err)
// 			errs = append(errs, err)
// 		}
// 		bytesFreed += size
// 	}
// 	if len(errs) > 0 {
// 		return Size(bytesFreed), fmt.Errorf("errors while removing files. errors: %s", errs)
// 	}
// 	h.l.Info("CacheHouseKeeping", "freed Bytes", bytesFreed)
// 	return Size(bytesFreed), nil
// }

// type ImageHandlerInfo struct {
// 	NumOfOriginals int
// 	NumOfCached    int
// 	OriginalsSize  Size
// 	CachedSize     Size
// }

// func (info ImageHandlerInfo) String() string {
// 	return fmt.Sprintf(`
// ImageHandlerInfo
// 	NumOfOriginals %d
// 	NumOfCached    %d
// 	OriginalsSize  %s
// 	CachedSize     %s
// `,
// 		info.NumOfOriginals,
// 		info.NumOfCached,
// 		info.OriginalsSize,
// 		info.CachedSize)
// }

// func (h *ImageHandler) Info() ImageHandlerInfo {
// 	oIds, err := h.ListIds()
// 	numOrigs := len(oIds)
// 	if err != nil {
// 		h.l.Error("Info", "error", err)
// 		numOrigs = -1
// 	}

// 	return ImageHandlerInfo{
// 		NumOfOriginals: numOrigs,
// 		NumOfCached:    h.cache.numberOfObjects,
// 		OriginalsSize:  Size(0),
// 		CachedSize:     h.cache.size,
// 	}
// }

func (h *ImageHandler) findLatestId() (int, error) {
	ids, err := h.ListIds()
	if err != nil {
		return 0, err
	}

	max := 0
	for _, id := range ids {
		if id > max {
			max = id
		}
	}
	return max, nil
}

// Create a new image with the given configuration and
// returns the path to the cached image.
func (h *ImageHandler) createImage(params ImageParameters, id int, cachePath string) (Size, error) {
	oPath := h.originalPath(id)
	oImg, err := loadImage(oPath)
	if err != nil {
		return 0, err
	}

	file, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	img := resize.Resize(params.Width, params.Height, oImg, resize.Lanczos3)

	switch params.Format {
	case Jpeg:
		if params.Quality == 0 {
			params.Quality = 80
		}
		opt := &jpeg.Options{Quality: params.Quality}
		err = jpeg.Encode(file, img, opt)
		if err != nil {
			return 0, err
		}
	case Png:
		err = png.Encode(file, img)
		if err != nil {
			return 0, err
		}
	case Gif:
		if params.Quality == 0 {
			params.Quality = 256
		}
		nc := params.Quality
		opt := &gif.Options{NumColors: nc}
		err = gif.Encode(file, img, opt)
		if err != nil {
			return 0, err
		}
	}
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	size := Size(stat.Size())
	if size == 0 {
		h.l.Error("createImage", "error", "created image has size "+size.String(), "path", cachePath)
		return 0, fmt.Errorf("created image has size 0")
	}
	return size, nil
}

func (h *ImageHandler) originalPath(id int) string {
	idStr := strconv.Itoa(id)
	return filepath.Join(h.conf.OriginalsDir, idStr+originalsExt)
}

func (h *ImageHandler) cachePath(params ImageParameters, id int) string {
	idStr := strconv.Itoa(id)
	return filepath.Join(h.conf.CacheDir, idStr+"_"+params.String())
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

// Represents file sizes:
//   - Infinite = 0
//   - 1 Kilobyte = 1024 bytes
//   - 1 Megabyte = 1024 Kilobytes
//   - 1 Gigabyte = 1024 Megabytes
//   - 1 Terabyte = 1024 Gigabytes
//   - 1 Petabyte = 1024 Terabytes
type Size uint64

const (
	Kilobyte = 1024            // 1 Kilobyte = 1024 bytes
	Megabyte = 1024 * Kilobyte // 1 Megabyte = 1024 Kilobytes
	Gigabyte = 1024 * Megabyte // 1 Gigabyte = 1024 Megabytes
	Terabyte = 1024 * Gigabyte // 1 Terabyte = 1024 Gigabytes
	Petabyte = 1024 * Terabyte // 1 Petabyte = 1024 Terabytes
)

// TODO: Break out size into a units package
// Size.String()
func (s Size) String() string {
	if s == 0 {
		return "0 B"
	}
	if s >= Petabyte {
		return sizeStringHelper(Petabyte, Terabyte, s, " PB")
	}
	if s >= Terabyte {
		return sizeStringHelper(Terabyte, Gigabyte, s, " TB")
	}
	if s >= Gigabyte {
		return sizeStringHelper(Gigabyte, Megabyte, s, " GB")
	}
	if s >= Megabyte {
		return sizeStringHelper(Megabyte, Kilobyte, s, " MB")
	}
	if s >= Kilobyte {
		return sizeStringHelper(Kilobyte, 1, s, " KB")
	}
	return fmt.Sprintf("%d B", s)

}

// sizeStringHelper returnes a string representing the size with two decimals.
// If the amount is exact then it will not show any decimals.
func sizeStringHelper(divInteger, divRemainder, s Size, suffix string) string {
	if s%divInteger == 0 {
		return fmt.Sprintf("%d%s", s/divInteger, suffix)
	}

	whole := s / divInteger
	rem := (s % divInteger) / divRemainder
	return signi3(int(whole), int(rem)) + suffix
}

// signi3 rounds to three significant digits given wholes and thousands
func signi3(whole, remainder int) string {
	if remainder < 5 {
		return fmt.Sprintf("%d.00", whole)
	}

	if remainder%10 >= 5 {
		remainder = (remainder + 5) / 10
	} else {
		remainder = remainder / 10
	}
	return fmt.Sprintf("%d.%d", whole, remainder)
}

func fileRemover(l *log.Logger, pathChan <-chan string) {
	l = l.WithPrefix("[FileRemover]")
	l.Debug("Starting file remover")
	for path := range pathChan {
		err := os.Remove(path)
		if err != nil {
			l.Error("Failed to remove file: ", err)
		}
		l.Debug("File removed", "path", path)
	}
	l.Error("File Remover stopped. Channel closed.")
}
