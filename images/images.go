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
	"sync"

	"github.com/charmbracelet/log"

	"github.com/nfnt/resize"
	//github.com/hashicorp/golang-lru/v2
	// TODO: consider using hashicorp lru or ARC for cache
	// build my own cache will be a good exercise. If I match the interface of hashicorp lru,
	// I can easily switch to it later. Or even let the user of this package give me a lru
	// eqvivalent on creation.
)

const (
	originalsExt = ".jpeg" //somewhat of a hack. all originals are saved and retrieved with this extention TODO: find a better way?
)

// ImageHandler is the main type of this package.
type ImageHandler struct {
	opts options

	mu       sync.Mutex
	latestId int

	cache cache
}

type ImageParameters struct {
	Id int

	Format Format

	// Jpeg:1-100, Gif:1-256
	Quality int

	// width and Height in pixels (0 = keep aspect ratio, both width and height can't be 0)
	Width  uint
	Height uint

	// Max file-size in bytes (0 = no limit)
	MaxSize Size
}

// New creates a new Imageandler and applies the given options.
//
// TODO: MUST create a cache based on files in cache folder on startup
func New(optMods ...optFunc) (*ImageHandler, error) {
	// set opts
	opts := optionsDefault()

	for _, fn := range optMods {
		err := fn(opts)
		if err != nil {
			opts.l.Fatal("Error setting options", "error", err)
			return nil, err
		}
	}

	l := opts.l
	l.Debug("Creating new ImageHandler", "number of options set", len(optMods), "resulting options", opts)

	err := checkDirs(opts)
	if err != nil {
		opts.l.Fatal("Error checking directories", "error", err)
		return nil, err
	}

	evitedChan := make(chan string, 10)
	go fileRemover(opts.l.WithPrefix("[file remover]"), evitedChan)

	ih := ImageHandler{
		opts: *opts,

		mu:       sync.Mutex{},
		latestId: 0,

		cache: NewLru(opts.cacheMaxNum, evitedChan),
	}

	ih.latestId, err = ih.findLatestId()
	if err != nil {
		opts.l.Fatal("could not get latest id during setup.", "error", err)
		return nil, err
	}
	return &ih, nil
}

// returns the path to the processed image.
func (h *ImageHandler) Get(params ImageParameters) (string, error) {
	// normalize parameters with defaults
	params.apply(h.opts.imageDefaults) //TODO: test this
	cachePath := h.cachePath(params)
	h.opts.l.Info("Get", "ImageParameters", params, "cachePath", cachePath)

	// Look for the image in the cache, return it if it does
	if inCache := h.cache.Access(cachePath); inCache {
		return cachePath, nil
	}

	// file does not exist
	size, err := h.createImage(params, params.Id, cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrIdNotFound{IdGiven: params.Id, Err: err}
		}
		return "", err
	}

	// file was created
	h.opts.l.Debug("Cachefile Created", "path", cachePath, "size", size)
	return cachePath, nil
}

func (h *ImageHandler) Add(path string) (int, error) {
	h.opts.l.Info("Add", "path", path)

	// check if file exists
	srcf, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("could not open source file: %w", err)
	}
	defer srcf.Close()

	// TODO: check with the image package instead of just the extension?
	// _, _, err = image.Decode(srcf)
	// if err != nil {
	// 	h.opts.l.Error("Error decoding image", "file", srcf, "error", err)
	// 	return 0, err
	// }

	// get a new id
	h.mu.Lock()
	h.latestId++
	id := h.latestId
	h.mu.Unlock()

	dst := h.opts.originalsDir + "/" + strconv.Itoa(id) + originalsExt
	// copy file to originals
	dstf, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, fmt.Errorf("could not open destination file: %w", err)
	}
	defer dstf.Close()
	_, err = io.Copy(dstf, srcf)
	if err != nil {
		return 0, fmt.Errorf("could not copy file: %w", err)
	}
	// return id
	return id, nil
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
	h.opts.l.Debug("ListIds")

	dir, err := os.Open(h.opts.originalsDir)
	if err != nil {
		h.opts.l.Error("ListIds", "error", err)
		return nil, err
	}
	defer dir.Close()

	// get list of files
	files, err := dir.Readdir(0)
	if err != nil {
		h.opts.l.Error("ListIds", "error", err)
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
			h.opts.l.Error("ListIds", "error", err)
			return nil, err
		}
		id := idInt

		h.opts.l.Debug("ListIds got", "id", id, "from", f.Name())
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
		h.opts.l.Error("createImage", "error", "created image has size "+size.String(), "path", cachePath)
		return 0, fmt.Errorf("created image has size 0")
	}
	return size, nil
}

func (h *ImageHandler) originalPath(id int) string {
	idStr := strconv.Itoa(id)
	return filepath.Join(h.opts.originalsDir, idStr+originalsExt)
}

func (h *ImageHandler) cachePath(params ImageParameters) string {
	return filepath.Join(h.opts.cacheDir, params.String())
}

// HELPER

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

func checkDirs(o *options) error {
	paths := []string{o.originalsDir, o.cacheDir}
	err := checkExists(paths, o.createDirs)
	if err != nil {
		return err
	}
	return checkFilePermissions(paths, o.setPermissions)
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

func MustSizeParse(str string) Size {
	s, err := SizeParse(str)
	if err != nil {
		panic(err)
	}
	return s
}

func SizeParse(str string) (Size, error) {
	if s, ok := sizeParseHelper(str, "KB", Kilobyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "MB", Megabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "GB", Gigabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "TB", Terabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "PB", Petabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "B", 1); ok {
		return s, nil
	}

	// no suffix, assume bytes
	if intVal, err := strconv.Atoi(str); err == nil && intVal >= 0 {
		return Size(intVal), nil
	}

	return Size(0), fmt.Errorf("could not parse size string '%s'", str)
}

func sizeParseHelper(str string, sufix string, size Size) (Size, bool) {
	before, found := strings.CutSuffix(str, sufix)
	if found {
		before = strings.Trim(before, " ")
		intVal, err := strconv.Atoi(before)
		if err != nil || intVal < 0 {
			return Size(0), false
		}
		return Size(intVal) * size, true
	}
	return Size(0), false
}

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

// TYPES and type parsers

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

func FormatParse(s string) (Format, error) {
	switch s {
	case "jpeg":
		return Jpeg, nil
	case "jpg":
		return Jpeg, nil
	case "png":
		return Png, nil
	case "gif":
		return Gif, nil
	}
	return "", fmt.Errorf("invalid image-format: %s", s)
}

func MustFormatParse(s string) Format {
	f, err := FormatParse(s)
	if err != nil {
		panic(err)
	}
	return f
}

// Interpolation represents interpolation methods used when resizing images.
type Interpolation string

const (
	NearestNeighbor Interpolation = "nearestNeighbor"
	Bilinear        Interpolation = "bilinear"
	Bicubic         Interpolation = "bicubic"
	MitchellNetrav  Interpolation = "MitchellNetravali"
	Lanczos2        Interpolation = "lanczos2"
	Lanczos3        Interpolation = "lanczos3"
)

func (r Interpolation) String() string {
	return string(r)
}

func ResizeInterpolationParse(s string) (Interpolation, error) {
	switch s {
	case "nearestNeighbor":
		return NearestNeighbor, nil
	case "bilinear":
		return Bilinear, nil
	case "bicubic":
		return Bicubic, nil
	case "MitchellNetravali":
		return MitchellNetrav, nil
	case "lanczos2":
		return Lanczos2, nil
	case "lanczos3":
		return Lanczos3, nil
	}
	return "", fmt.Errorf("invalid resize-interpolation-function: %s", s)
}

func MustResizeInterpolationParse(s string) Interpolation {
	r, err := ResizeInterpolationParse(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (ip *ImageParameters) String() string {
	return fmt.Sprintf("%d_%dx%d_q%d_s%d.%s", ip.Id, ip.Width, ip.Height, ip.Quality, ip.MaxSize, ip.Format)
}

func (ip *ImageParameters) apply(def ImageDefaults) {
	if ip.Format == "" {
		ip.Format = def.format
	}
	if ip.Quality == 0 && ip.Format == Jpeg {
		ip.Quality = def.qualityJpeg
	}
	if ip.Quality == 0 && ip.Format == Gif {
		ip.Quality = def.qualityGif
	}
	if ip.Width == 0 && ip.Height == 0 {
		ip.Width = uint(def.width)
		ip.Height = uint(def.height)
	}
	if ip.MaxSize == 0 {
		ip.MaxSize = def.maxSize
	}
}

// cache is expected to be thread-safe. It should return true if the path is in cache and false otherwise.
//
// It should add the path to cache if it is not already there.
//
// NOTE: cache should send evicted paths through a channel to the image handler to be deleted from disk.
type cache interface {
	Access(path string) bool
}

// CONFIGURATION

// options represents the configuration options for the ImageHandler
type options struct {
	l        *log.Logger
	logLevel log.Level

	createDirs     bool
	setPermissions bool

	originalsDir string

	cacheDir     string
	cacheMaxNum  int
	cacheMaxSize Size

	imageDefaults ImageDefaults
	imagePresets  []ImagePreset

	interpolation Interpolation
}

func (o *options) String() string {
	strB := strings.Builder{}
	strB.WriteString("Options:\n")
	strB.WriteString(fmt.Sprintf("  logLevel: %s\n", o.logLevel))
	strB.WriteString(fmt.Sprintf("  createDirs: %t\n", o.createDirs))
	strB.WriteString(fmt.Sprintf("  setPermissions: %t\n", o.setPermissions))
	strB.WriteString(fmt.Sprintf("  originalsDir: %s\n", o.originalsDir))
	strB.WriteString(fmt.Sprintf("  cacheDir: %s\n", o.cacheDir))
	strB.WriteString(fmt.Sprintf("  cacheMaxNum: %d\n", o.cacheMaxNum))
	strB.WriteString(fmt.Sprintf("  cacheMaxSize: %s\n", o.cacheMaxSize))
	strB.WriteString(fmt.Sprintf("  interpolation: %s\n", o.interpolation))
	strB.WriteString(fmt.Sprintf("  imageDefaults: %s\n", o.imageDefaults))
	strB.WriteString(fmt.Sprintf("  imagePresets: %s\n", o.imagePresets))
	return strB.String()

}

type ImageDefaults struct {
	format      Format
	qualityJpeg int
	qualityGif  int
	width       int
	height      int
	maxSize     Size
}

func (id ImageDefaults) String() string {
	strB := strings.Builder{}
	strB.WriteString("\n")
	strB.WriteString(fmt.Sprintf("    format: %s\n", id.format))
	strB.WriteString(fmt.Sprintf("    qualityJpeg: %d\n", id.qualityJpeg))
	strB.WriteString(fmt.Sprintf("    qualityGif: %d\n", id.qualityGif))
	strB.WriteString(fmt.Sprintf("    width: %d\n", id.width))
	strB.WriteString(fmt.Sprintf("    height: %d\n", id.height))
	strB.WriteString(fmt.Sprintf("    maxSize: %s", id.maxSize))
	return strB.String()

}

type ImagePreset struct {
	Name    string
	Alias   []string
	Format  Format
	Quality int
	Width   int
	Height  int
	MaxSize Size
}

func (ip ImagePreset) String() string {
	strB := strings.Builder{}
	strB.WriteString("\n")
	strB.WriteString(fmt.Sprintf("    %s:\n", ip.Name))
	strB.WriteString(fmt.Sprintf("      alias: %s\n", ip.Alias))
	strB.WriteString(fmt.Sprintf("      format: %s\n", ip.Format))
	strB.WriteString(fmt.Sprintf("      quality: %d\n", ip.Quality))
	strB.WriteString(fmt.Sprintf("      width: %d\n", ip.Width))
	strB.WriteString(fmt.Sprintf("      height: %d\n", ip.Height))
	strB.WriteString(fmt.Sprintf("      maxSize: %s", ip.MaxSize))
	return strB.String()
}

func optionsDefault() *options {
	return &options{
		l:        log.Default(),
		logLevel: log.InfoLevel,

		createDirs:     false,
		setPermissions: false,

		originalsDir: "img/originals",

		cacheDir:     "img/cache",
		cacheMaxNum:  1000000,
		cacheMaxSize: 10 * Gigabyte,

		imageDefaults: ImageDefaults{
			format:      Jpeg,
			qualityJpeg: 80,
			qualityGif:  256,
			width:       0,
			height:      800,
			maxSize:     10 * Megabyte,
		},

		imagePresets: []ImagePreset{
			{Name: "thumb", Alias: []string{"t", "th", "thumb"}, Format: Jpeg, Quality: 80, Width: 0, Height: 200, MaxSize: 1 * Megabyte},
			{Name: "small", Alias: []string{"s", "small"}, Format: Jpeg, Quality: 80, Width: 0, Height: 400, MaxSize: 2 * Megabyte},
			{Name: "medium", Alias: []string{"m", "medium"}, Format: Jpeg, Quality: 80, Width: 0, Height: 800, MaxSize: 4 * Megabyte},
			{Name: "large", Alias: []string{"l", "large"}, Format: Jpeg, Quality: 80, Width: 0, Height: 1600, MaxSize: 8 * Megabyte},
			{Name: "hero", Alias: []string{"xl", "hero"}, Format: Jpeg, Quality: 80, Width: 0, Height: 3200, MaxSize: 16 * Megabyte},
		},

		interpolation: NearestNeighbor,
	}
}

// optFunc is a function for setting an option
type optFunc func(*options) error

// WithLogger sets the logger
func WithLogger(l *log.Logger) optFunc {
	return func(o *options) error {
		o.l = l
		return nil
	}
}

// WithLogLevel sets the log level
func WithLogLevel(s string) optFunc {
	return func(o *options) error {
		l := log.ParseLevel(s)
		o.logLevel = l
		return nil
	}
}

// WithCreateDirs sets the create directories option
func WithCreateDirs(o *options) error {
	o.createDirs = true
	return nil
}

// WithSetPermissions sets the set permissions option
func WithSetPermissions(o *options) error {
	o.setPermissions = true
	return nil
}

// WithOriginalsDir sets the originals directory
func WithOriginalsDir(dir string) optFunc {
	return func(o *options) error {
		o.originalsDir = dir
		return nil
	}
}

// WithCacheDir sets the cache directory
func WithCacheDir(dir string) optFunc {
	return func(o *options) error {
		o.cacheDir = dir
		return nil
	}
}

// WithCacheMaxNum sets the cache max number option
func WithCacheMaxNum(num int) optFunc {
	return func(o *options) error {
		o.cacheMaxNum = num
		return nil
	}
}

// WithCacheMaxSize sets the cache max size option
func WithCacheMaxSize(size Size) optFunc {
	return func(o *options) error {
		o.cacheMaxSize = size
		return nil
	}
}

// WithImageDefaults sets defaults used when no preset or parameters are given
func WithImageDefaults(id ImageDefaults) optFunc {
	return func(o *options) error {
		o.imageDefaults = id
		return nil
	}
}

// WithImagePreset adds a preset to the handler
func WithImagePreset(preset ImagePreset) optFunc {
	return func(o *options) error {
		o.imagePresets = append(o.imagePresets, preset)
		return nil
	}
}

// WithInterpolation sets the interpolation method used when resizing images
func WithInterpolation(i Interpolation) optFunc {
	return func(o *options) error {
		o.interpolation = i
		return nil
	}
}

// ERRORS

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
