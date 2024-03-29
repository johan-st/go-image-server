package images

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/units/size"

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

	presets map[string]ImagePreset
}

type Stat struct {
	Ids      []int
	SizeOrig size.S
	Cache    CacheStat
}

type ImageStat struct {
	Id           int
	OriginalSize size.S
	CacheSize    size.S
	CacheNum     int
}

type ImageParameters struct {
	Id int

	Format

	// Jpeg:1-100, Gif:1-256
	Quality int

	// width and Height in pixels (0 = keep aspect ratio, both width and height can not be 0)
	Width  uint
	Height uint

	// Max file-size in bytes (0 = no limit)
	MaxSize size.S

	// TODO: implement
	// Interpolation function used if a new cache file is created
	// Interpolation Interpolation
}

// New creates a new Imageandler and applies the given options.
//
// TODO: MUST create a cache based on files in cache folder on startup
func New(optFuncs ...optFunc) (*ImageHandler, error) {
	// set opts
	opts := optionsDefault()

	for _, fn := range optFuncs {
		err := fn(&opts)
		if err != nil {
			return nil, err

		}
	}

	// If no logger was set. Discard all logging from package.¨
	// TODO: build noop logger or disable logging in a way that require less allocs
	if opts.l == nil {
		opts.l = log.New(io.Discard)
	}

	l := opts.l

	err := checkDirs(l, &opts)
	if err != nil {
		return nil, err
	}

	evictedChan := make(chan string, 128)
	go fileRemover(opts.l.WithPrefix("[file remover]"), evictedChan)

	ih := ImageHandler{
		opts: opts,

		mu:       sync.Mutex{},
		latestId: 0,

		cache: newLru(opts.cacheMaxNum, evictedChan),

		presets: presetsMap(opts.imagePresets),
	}

	ih.latestId, err = ih.findLatestId()
	if err != nil {
		opts.l.Fatal("could not get latest id during setup.", "error", err)
		return nil, err
	}

	l.Debug("Creating new ImageHandler", "number of options set", len(optFuncs), "resulting options", opts.String())
	return &ih, nil
}

// returns the path to the processed image.
func (h *ImageHandler) Get(params ImageParameters) (string, error) {
	// normalize parameters with defaults
	params.apply(h.opts.imageDefaults) //TODO: test this
	cachePath := h.cachePath(params)
	h.opts.l.Debug("Get", "ImageParameters", params, "cachePath", cachePath)

	// Look for the image in the cache, return it if it does
	if inCache := h.cache.Contains(cachePath); inCache {
		h.cache.AddOrUpdate(params.Id, cachePath)
		return cachePath, nil
	}

	// file does not exist
	size, err := h.createImage(params, cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrIdNotFound{IdGiven: params.Id, Err: err}
		}
		return "", err
	}

	// file was created, adding to cache.
	h.cache.AddOrUpdate(params.Id, cachePath)
	h.opts.l.Debug("Cachefile Created", "path", cachePath, "size", size)
	return cachePath, nil
}

// Returns id of the added image
func (h *ImageHandler) Add(r io.Reader) (int, error) {
	h.opts.l.Debug("Add called on imageHandler")

	// temp file
	tmpFile, err := os.CreateTemp(h.opts.dirCache, "upload-*")
	if err != nil {
		return 0, fmt.Errorf("could not create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	tr := io.TeeReader(r, tmpFile)

	// decode image
	_, _, err = image.Decode(tr)
	if err != nil {
		return 0, err
	}

	// seek to start
	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("could not read from tmpFile: %w", err)
	}

	// get a new id
	h.mu.Lock()
	h.latestId++
	id := h.latestId
	h.mu.Unlock()

	dst := h.opts.dirOriginals + "/" + strconv.Itoa(id) + originalsExt
	// copy file to originals
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, fmt.Errorf("could not open destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, tmpFile)
	if err != nil {
		return 0, fmt.Errorf("could not copy file: %w", err)
	}

	// return id
	return id, nil
}

// TODO: page and chunk as arguments for when we have thousands of ids?
func (h *ImageHandler) Ids() ([]int, error) {
	h.opts.l.Debug("ListIds")

	dir, err := os.Open(h.opts.dirOriginals)
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

		ids = append(ids, id)
	}
	return ids, nil
}

func (h *ImageHandler) GetPreset(preset string) (ImagePreset, bool) {
	p, ok := h.presets[preset]
	if !ok {
		return ImagePreset{}, false
	}
	return p, true
}

func (h *ImageHandler) Delete(id int) error {
	h.opts.l.Debug("Delete", "id", id)
	// TODO: lock while deleting?
	oPath := h.originalPath(id)
	err := os.Remove(oPath)
	if err != nil {
		return err
	}

	numDeleted := h.cache.Delete(id)
	h.opts.l.Debug("Delete", "cache entries removed", numDeleted)

	return nil
}

func (h *ImageHandler) Stat() (Stat, error) {

	ids, err := h.Ids()
	if err != nil {
		return Stat{}, err
	}
	var sizeOrig size.S
	for _, i := range ids {
		size, err := sizeFile(h.originalPath(i))
		if err != nil {
			return Stat{}, err
		}
		sizeOrig += size
	}
	return Stat{
		Ids:      ids,
		SizeOrig: sizeOrig,
		Cache:    h.cache.Stat(),
	}, err
}

func (h *ImageHandler) StatId(id int) (ImageStat, error) {
	h.opts.l.Debug("StatId", "id", id)

	oPath := h.originalPath(id)
	oSize, err := sizeFile(oPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ImageStat{}, ErrIdNotFound{IdGiven: id, Err: err}
		}
		return ImageStat{}, err
	}

	cPaths := h.cache.Get(id)
	cSize := size.S(0)
	for _, p := range cPaths {
		size, err := sizeFile(p)
		if err != nil {
			return ImageStat{}, err
		}
		cSize += size
	}

	return ImageStat{
		Id:           id,
		OriginalSize: oSize,
		CacheSize:    cSize,
		CacheNum:     len(cPaths),
	}, nil
}
func sizeFile(p string) (size.S, error) {
	info, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return size.S(info.Size()), nil
}

func (h *ImageHandler) findLatestId() (int, error) {
	ids, err := h.Ids()
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
func (h *ImageHandler) createImage(params ImageParameters, cachePath string) (size.S, error) {
	oPath := h.originalPath(params.Id)
	oImg, err := loadImage(oPath)
	if err != nil {
		return 0, err
	}

	file, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	if params.Width != 0 && params.Height != 0 {
		oImg = cropToRatio(oImg, int(params.Width), int(params.Height))
	}

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
	size := size.S(stat.Size())
	if size == 0 {
		h.opts.l.Error("createImage", "error", "created image has size.size "+size.String(), "path", cachePath)
		return 0, fmt.Errorf("created image has size.size 0")
	}
	return size, nil
}

func (h *ImageHandler) originalPath(id int) string {
	idStr := strconv.Itoa(id)
	return filepath.Join(h.opts.dirOriginals, idStr+originalsExt)
}

func (h *ImageHandler) cachePath(params ImageParameters) string {
	return filepath.Join(h.opts.dirCache, params.String())
}

// HELPER

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

// TODO: IMPLEMENT
func cropToRatio(i image.Image, width int, height int) image.Image {
	dx := i.Bounds().Dx()
	dy := i.Bounds().Dy()
	r := float64(width) / float64(height)
	w, h := ratioToPixels(
		r,
		float64(dx),
		float64(dy),
	)
	cropX := (dx - w) / 2
	cropY := (dy - h) / 2

	subRect := image.Rect(cropX, cropY, w+cropX, h+cropY)
	if subI, ok := i.(SubImager); ok {
		return subI.SubImage(subRect)
	} else {
		// TODO: need to be looked over when we add imagetypes..
		// Assumption is tha we do not use any format that does not implement SubImages
		panic("image was not of type SubImager")
	}
}

func ratioToPixels(ratioWH, width, height float64) (w int, h int) {
	if width < 1 || height < 1 || ratioWH <= 0 {
		return 0, 0
	}
	if height*ratioWH <= width {
		return int(math.Round(height * ratioWH)), int(math.Round(height))
	} else {
		return int(math.Round(width)), int(math.Round(width / ratioWH))
	}
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

func checkDirs(l *log.Logger, o *options) error {
	paths := []string{o.dirOriginals, o.dirCache}
	err := checkExists(l, paths, o.createDirs)
	if err != nil {
		return err
	}
	return checkFilePermissions(l, paths, o.setPermissions)
}

func checkExists(l *log.Logger, paths []string, createDirs bool) error {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if createDirs {
				l.Info("creating directory", "path", path)
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
func checkFilePermissions(l *log.Logger, paths []string, setPerms bool) error {
	for _, path := range paths {
		err := filepath.WalkDir(path, permAtLeast(l, 0700, 0600))
		if err != nil {
			return err
		}
	}
	return nil
}

// Will extend permissions if needed, but will not reduce them.
func permAtLeast(l *log.Logger, dir os.FileMode, file os.FileMode) fs.WalkDirFunc {
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

		// bitwise and on
		if i.Mode().Perm()&perm < perm {
			p := i.Mode().Perm() | perm
			err := os.Chmod(path, p)
			if err != nil {
				return fmt.Errorf("'%s' has insufficient permissions and setting new permission failed. (was: %o, need at least: %o)", path, i.Mode().Perm(), perm)
			}
			l.Warn("insufficient permissions", "path", path, "was", i.Mode().Perm(), "need at least", perm, "set to", p)

		}
		return nil
	}
}

func fileRemover(l *log.Logger, pathChan <-chan string) {
	l.Debug("Running...")
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

func ParseFormat(s string) (Format, error) {
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
	return "", fmt.Errorf("invalid image-format. \n\tGot: %s\n\tWant: 'jpeg', 'jpg', 'png', 'gif'", s)
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

func ParseInterpolation(s string) (Interpolation, error) {
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
	return "", fmt.Errorf("invalid interpolation-function. \n\tGot: '%s'\n\tWant: 'nearestNeighbor', 'bilinear', 'bicubic', 'MitchellNetravali', 'lanczos2', 'lanczos3'", s)
}

func (ip *ImageParameters) String() string {
	return fmt.Sprintf("%d_%dx%d_q%d_s%d.%s", ip.Id, ip.Width, ip.Height, ip.Quality, ip.MaxSize, ip.Format)
}

func (ip *ImageParameters) apply(def ImageDefaults) {
	if ip.Format == "" {
		ip.Format = def.Format
	}
	if ip.Quality == 0 && ip.Format == Jpeg {
		ip.Quality = def.QualityJpeg
	}
	if ip.Quality == 0 && ip.Format == Gif {
		ip.Quality = def.QualityGif
	}
	if ip.Width == 0 && ip.Height == 0 {
		ip.Width = uint(def.Width)
		ip.Height = uint(def.Height)
	}
	if ip.MaxSize == 0 {
		ip.MaxSize = def.MaxSize
	}
}

// cache is expected to be thread-safe.
//
// NOTE: cache should send evicted paths through a channel to the image handler to be deleted from disk.
type cache interface {
	Contains(path string) bool
	AddOrUpdate(id int, path string) bool
	Delete(id int) int
	Stat() CacheStat
	Get(id int) []string //returns a slice of paths to cached images with given ID
}

type CacheStat struct {
	NumItems  int
	Capacity  int
	Size      size.S
	Hit       uint32
	Miss      uint32
	Evictions uint32
}

// CONFIGURATION

// options represents the configuration options for the ImageHandler
type options struct {
	l        *log.Logger
	logLevel log.Level

	createDirs     bool
	setPermissions bool

	dirOriginals string
	dirCache     string

	cacheMaxNum  int
	cacheMaxSize size.S

	imageDefaults ImageDefaults
	imagePresets  []ImagePreset
}

func (o *options) String() string {
	strB := strings.Builder{}
	strB.WriteString("Options:\n")
	if o.l != nil {
		strB.WriteString("  logging: yes\n")
		strB.WriteString(fmt.Sprintf("  logLevel: %s\n", o.logLevel))
	} else {
		strB.WriteString("  logging: no\n")
	}
	strB.WriteString(fmt.Sprintf("  createDirs: %t\n", o.createDirs))
	strB.WriteString(fmt.Sprintf("  setPermissions: %t\n", o.setPermissions))
	strB.WriteString(fmt.Sprintf("  originalsDir: %s\n", o.dirOriginals))
	strB.WriteString(fmt.Sprintf("  cacheDir: %s\n", o.dirCache))
	strB.WriteString(fmt.Sprintf("  cacheMaxNum: %d\n", o.cacheMaxNum))
	strB.WriteString(fmt.Sprintf("  cacheMaxSize: %s\n", o.cacheMaxSize))
	strB.WriteString(fmt.Sprintf("  imageDefaults: %s\n", o.imageDefaults))
	strB.WriteString(fmt.Sprintf("  imagePresets: %+v\n", o.imagePresets))
	return strB.String()

}

type ImageDefaults struct {
	Format      Format
	QualityJpeg int
	QualityGif  int
	Width       int
	Height      int
	MaxSize     size.S
	Interpolation
}

func (id ImageDefaults) String() string {
	strB := strings.Builder{}
	strB.WriteString("\n")
	strB.WriteString(fmt.Sprintf("    format: %s\n", id.Format))
	strB.WriteString(fmt.Sprintf("    qualityJpeg: %d\n", id.QualityJpeg))
	strB.WriteString(fmt.Sprintf("    qualityGif: %d\n", id.QualityGif))
	strB.WriteString(fmt.Sprintf("    width: %d\n", id.Width))
	strB.WriteString(fmt.Sprintf("    height: %d\n", id.Height))
	strB.WriteString(fmt.Sprintf("    maxSize: %s\n", id.MaxSize))
	strB.WriteString(fmt.Sprintf("    interpolation: %s", id.Interpolation))
	return strB.String()
}

func (id *ImageDefaults) validate() error {
	return nil
	// TODO: this function
}

type ImagePreset struct {
	Name    string
	Alias   []string
	Format  Format
	Quality int
	Width   int
	Height  int
	MaxSize size.S
	Interpolation
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
	strB.WriteString(fmt.Sprintf("      maxSize: %s\n", ip.MaxSize))
	strB.WriteString(fmt.Sprintf("      interpolation: %s", ip.Interpolation))
	return strB.String()
}

func (ip ImagePreset) validate() error {
	return nil
	// TODO: this function
}

func optionsDefault() options {
	return options{
		l:        nil,
		logLevel: log.InfoLevel,

		createDirs:     false,
		setPermissions: false,

		dirOriginals: "img/originals",

		dirCache:     "img/cache",
		cacheMaxNum:  1000000,
		cacheMaxSize: 10 * size.Gigabyte,

		imageDefaults: ImageDefaults{
			Format:      Jpeg,
			QualityJpeg: 80,
			QualityGif:  256,
			Width:       0,
			Height:      800,
			MaxSize:     10 * size.Megabyte,
		},

		imagePresets: []ImagePreset{},
	}
}

func presetsMap(imagePresets []ImagePreset) map[string]ImagePreset {
	m := make(map[string]ImagePreset)
	for _, p := range imagePresets {
		for _, a := range p.Alias {
			m[a] = p
		}
	}
	return m
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
		o.l.SetLevel(l)
		return nil
	}
}

// WithCreateDirs sets the create directories option
func WithCreateDirs(b bool) optFunc {
	return func(o *options) error {
		o.createDirs = b
		return nil
	}
}

// WithSetPermissions sets the set permissions option
func WithSetPermissions(b bool) optFunc {
	return func(o *options) error {
		o.setPermissions = b
		return nil
	}
}

// WithOriginalsDir sets the originals directory
func WithOriginalsDir(dir string) optFunc {
	return func(o *options) error {
		o.dirOriginals = dir
		return nil
	}
}

// WithCacheDir sets the cache directory
func WithCacheDir(dir string) optFunc {
	return func(o *options) error {
		o.dirCache = dir
		return nil
	}
}

// WithCacheMaxNum sets the cache max number option
func WithCacheMaxNum(num int) optFunc {
	return func(o *options) error {
		if num < 1 {
			return fmt.Errorf("cache can not be smaller then 1 image. got: %d", num)
		}
		o.cacheMaxNum = num
		return nil
	}
}

// WithCacheMaxSize sets the cache max size.size option
func WithCacheMaxSize(size size.S) optFunc {
	return func(o *options) error {
		o.cacheMaxSize = size
		return nil
	}
}

// WithImageDefaults sets defaults used when no preset or parameters are given
func WithImageDefaults(id ImageDefaults) optFunc {
	return func(o *options) error {
		err := id.validate()
		if err != nil {
			return err
		}
		o.imageDefaults = id
		return nil
	}
}

// WithImagePresets adds the given set of presets to the handler
func WithImagePresets(presets []ImagePreset) optFunc {
	return func(o *options) error {
		errs := []error{}
		for _, p := range presets {
			errs = append(errs, p.validate())
			o.imagePresets = append(o.imagePresets, p)
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
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

// Logger

type Logger interface {
	Debug(msg any, keyvals ...any)
	Info(msg any, keyvals ...any)
	Warn(msg any, keyvals ...any)
	Error(msg any, keyvals ...any)
	Fatal(msg any, keyvals ...any)
}
