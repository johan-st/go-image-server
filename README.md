# go-image-server

one stop shop for serving the images you need in the formate you want.

Idea is to be able to upload a an image in high resolution and quality and have the server handle scaling etc (as per query parameters). The image server should cache images it modifies to speed up future delivery of the same request.

Idealy image workflow should be significantly simplified. Workflow could be to upload a single high-res version of each image to the image-server and let all teams fetch the sizes, qualities and formats they require for each application. Once a specific image has been requested it will be cached and instantly available on subsequent requests.

## Proof of concept

A proof of concept **MUST** be able to serve images in the requested pixel-size.
It **SHOULD** have at least two (2) quality levels and **SHOULD** be able to cache images.

## Minimum Viable Product

A MVP **MUST** be able to serve images in the requested pixel-size and **SHOULD** have rudimentary authentication/domain restrictions.

## Table Of Contents

- [go-image-server](#go-image-server)
  - [Proof of concept](#proof-of-concept)
  - [Minimum Viable Product](#minimum-viable-product)
  - [Table Of Contents](#table-of-contents)
    - [Project Requirements](#project-requirements)
      - [Server](#server)
  - [Sprints](#sprints)
    - [MoSCoW](#moscow)
    - [June 2023](#june-2023)
      - [backlog (june 2023)](#backlog-june-2023)
      - [log (june 2023)](#log-june-2023)
    - [April \& May 2023](#april--may-2023)
      - [backlog (april \& may 2023)](#backlog-april--may-2023)
      - [log (april \& may 2023)](#log-april--may-2023)
    - [Februari 2022](#februari-2022)
      - [backlog (februari 2022)](#backlog-februari-2022)
      - [log (februari 2022)](#log-februari-2022)
  - [Requirements](#requirements)
  - [Questions to answer](#questions-to-answer)
    - [structure of request](#structure-of-request)
    - [Caching](#caching)
  - [thoughts](#thoughts)
    - [images package API (DRAFT)](#images-package-api-draft)
    - [Performance](#performance)
  - [Known issues](#known-issues)

### Project Requirements

#### Server

Server **MUST** have sufficient authorization and domain restrictions to avoid abuse.
Server **SHOULD** have a mechanism for finding errors and possible abuse.
Server **MUST** be able to add images while running.
Server **MUST** be able to add images from a folder on startup.
Server **COULD** monitor a folder for new images and add them automatically.
Server **COULD** serve thumbnail of image while creating a new cached version.
Images requested **SHOULD** be served within 500ms on first request and **MUST** be served within 25ms on subsequent requests.
Images **MUST** be served in the size requested.
Images **MUST** not be stretched or distorted.
Images **SHOULD** not be scaled up from original.
Images **SHUOLD** have quality-options
Images **COULD** have option for added watermark
Images **COULD** have option for added text
Images **COULD** check for duplicates on upload (bloom filter?)
Images **COULD** use AI for creating tags and metadata such as alt-text, description, title etc
Images **COULD** have a search function for finding images based on tags, metadata etc
Code **MUST** have relevant tests
Code **SHOULD** be benchmarked for performance
Code **COULD** be profiled for cpu and memory usage
Code **SHOULD** be maintainable and easy to understand.

## Sprints

### [MoSCoW](https://en.wikipedia.org/wiki/MoSCoW_method)

Planing and prioritization of features and requirements to be implemented during each sprint.

Requirements not mentioned should be regarded as **WONT**
### July & August 2023
_summer break_

#### log (july & august 2023)
- 2023-08-16: bugfix for lru, add debug flag, update readme
- 2023-08-17: add dockerfile



### June 2023

- [x] **MUST** have prototype admin for uploading images
- [x] **MUST** have prototype admin for viewing and deleting images
- [X] **SHOULD** have prototype info page for viewing server status, uptime, cache size etc
- [x] **SHOULD** keep aspect ratio when cropping
- [x] **COULD** have prototype admin for viewing and deleting cached images
- [x] **COULD** be benchmarked for performance

#### backlog (june 2023)

- [ ] decide on how to handle handle requests for images larger than original?
- [x] crop should keep aspect ratio
- [X] handle folder creation and permissions
  - [ ] win
  - [ ] linux
- [ ] tests need to be able to run with no setup after clone (include sane default conf in repo)
- [ ] app need to be able to run as binary with no setup (maybe sane default conf is created?)
- [ ] support webp
- [ ] safe concurrency
- [x] switch to disable docs
- [ ] decide on pathing when calling from a different folder than the binary
- [ ] implement interpolation function
  - [ ] benchmark functions and make a note in the docs
- [ ] load images and cache from disk on startup
- [ ] investigate hardcoding docs into binary
- [ ] "Add folder" on startup should be reccursive?
- [ ] Strip images package to bare minimum. Move unnecessary stuff to main package
- [ ] update USAGE.md
- [ ] make the creation of a default config file when no conf was found an opt-in feature
- [ ] handle width and height from diferent sources (query params, presets, defaults)
- [ ] consider having a fallback image for when the requested image is not found
- [ ] consider having a fallback image or a smaller image for quick response when the requested image is not cached
- [ ] consider commiting to single executable? as of now I need docs folder and config file
- [ ] add cacnelation context to "add folder on startup" and "load images and cache from disk on startup"?
- [X] decide on benchmark-method for single images
- [X] asking for a non-exsistant id should give 404 (not 500)
- [x] add usage log
- [x] add cache reclaimation
- [ ] decide on how to handle handle requests for images larger than original?
- [x] crop should keep aspect ratio
- [ ] consider storing image-ids in a lookup table for quick "exists" checks.
- [ ] consider storing loaded images (`type Image` from image package) in memory for quick access when creating multiple new sizes
- [ ] consider storing image-files in memory for quicker responses (relly on config max or maybe use resource monitoring to not run out of memory?)
- [ ] log an info level message when deleting an image
- [ ] BUG: delete requests to a non-existant id causes a 500-response
  - [ ] response should probably be a 400
  - [ ] logging as error in handler. should be info or warn
- [ ] PROBLEM: benchmarks are broken...
- [ ] consider migrating docs to templates instead of markdown?
- [ ] refactor images package
- [ ] refuse duplicate uploads
- [ ] find similar images
- [ ] add metadata to images
- [ ] add tags to images
- [ ] searchable tags and metadata
- [ ] AI can create tags and metadata such as description, alt-text, title etc

#### log (june 2023)

- 2023-06-01: wip. refactoring logging and config. Banchmarks are a lot worse than before. Need to investigate.
- 2023-06-02: update README.md
- 2023-06-02: apply logging and config changes
- 2023-06-03: add docs to API
- 2023-06-03: set up endpoint for DELETE /images/:id
- 2023-06-04: bugfix for lru
- 2023-06-08: Learn template basics and create admin page prototype.
- 2023-06-09: update README.md
- 2023-06-09: crop instead of stretch images
- 2023-06-12: add info page and have lru keep basic stats
- 2023-06-12: fix bug with Size.String()
- 2023-06-13: update info and add info for specific image
- 2023-06-20: add openapi yaml definition
- 2023-06-20: refactor api/image -> api/images
- 2023-06-20: refactor images/size into units/size (package size)

### April & May 2023

(#backatit)

- [x] **MUST** have tests working and passing on linux and windows
  - [x] linux
  - [X] windows
- [x] **SHOULD** error on startup if file permissions are wrong
  - Decided to have the server try to set the permissions for image and cache folder itself if they are not permissive enough. It will only extend permissions, never reduce them.
- [x] **SHOULD** be configurable by file or environment variables
  - [ ] **SHOULD** be propperly integrated and tested
- [ ] **COULD** have a webpage for uploading images
- [ ] **COULD** have a webpage for viewing and deleting images

#### backlog (april & may 2023)

- [X] Fix tests
- [X] asking for a non-exsistant id should give 404 (not 500)
  - was already fixed in previous sprint
- [X] Paths should be configurable
- [X] Refactor tests to work on the image modules API (not as part of the module)  
- [X] decide on benchmark-method for single images
  - I will use the `go test -bench=. -benchmem` method
- [x] add usage log
- [X] add simple cache retention and reclaimation
- [ ] decide on how to handle handle requests for images larger than original?
- [ ] crop should keep aspect ratio
- [X] handle folder creation and permissions
  - [ ] win
  - [ ] linux
- [ ] tests need to be able to run with no setup after clone (include sane default conf in repo)
- [ ] app need to be able to run as binary with no setup (maybe sane default conf is created?)
- [ ] support webp
- [ ] safe concurrency
- [ ] switch to disable docs
- [x] consider using OptFunc pattern for config.
  - e.g: `NewServer(withPort(8080), withCacheSize(1000), withImageDefaults(imgDef), withImagePreset(imgPresetSmall), withImagePreset(imgPresetThumb))` etc...
  - [https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
  - Anthony GG video: [https://www.youtube.com/watch?v=MDy7JQN5MN4](https://www.youtube.com/watch?v=MDy7JQN5MN4)
- [X] Stricter checks when adding images. Should be imposible to add a file that can not be parsed as an image.
- [ ] decide on pathing when calling from a different folder than the binary
- [X] docs should be opt-in with a flag and config-file
- [X] 'clear cache on startup' should be 'clear cache on shutdown'
- [ ] inplement interpolation function
- [ ] load images and cache from disk on startup
- [ ] investigate hardcoding docs into binary
- [ ] Add folder on startup should be reccursive?
- [ ] Strip images package to bare minimum. Move unnecessary stuff to main package
- [ ] update USAGE.md
- [ ] make the creation of a default config file when no conf was found an opt-in feature
- [ ] handle width and height from diferent sources (query params, presets, defaults)
- [ ] consider having a fallback image for when the requested image is not found
- [ ] consider having a fallback image or a smaller image for quick response when the requested image is not cached
- [X] Access Log as midleware in router
- [ ] consider commiting to single executable? as of now I need docs folder and config file
- [ ] add cacnelation context to "add folder on startup" and "load images and cache from disk on startup"?

#### log (april & may 2023)

- 2023-04-25: look over code and tests. Plan future work
- 2023-04-26: Fix permission issues with cache and image folders. Tests now pass on linux
- 2023-04-26: slight restructure of README.md
- 2023-04-26: Fix tests (and image module) on windows
- 2023-04-26: refactor out io/util as it is deprecated
- 2023-04-28: Think through images package API (images_api_test.go)
- 2023-05-02: further work on types and interface for images package
- 2023-05-03: further work on types and interface for images package
- 2023-05-04: use charm.io log. continue refatoring...  
- 2023-05-05: implement "not implemented yet" func. there are a few left for work with caching  
- 2023-05-05: start implementing a cache backed by a slice. This was the simplest way I could think of while working on the cache API
- 2023-05-06: play with caching and trying to figure out what I want and need.
- 2023-05-07: update docs (USAGE.md) and restructure repo
- 2023-05-08: start caching experiments
- 2023-05-12: experimentation continues
- 2023-05-17: hook in bespoke LRU (least recently used) cache
- 2023-05-18: implement defaults for ImageHadler and ImageParameters
- 2023-05-18: set up basics for yaml config. this will supercede the defaults when implemented
- 2023-05-19: hook in yaml config. All settings are not yet used.
- 2023-05-19: create branch for working on a new config pattern for images package
- 2023-05-24: branch: conf another way. refactoring and hook in config file
- 2023-05-25: work on conf and update backlogs
- 2023-05-26: work on conf and check a few backlogs
- 2023-05-28: Fix bug and add access log as file
- 2023-05-31: create image upload endpoint.

### Februari 2022

- [x] **MUST** store and serve images
- [x] **SHOULD** preprocess images on demand
  - [x] **SHOULD** resize
  - [x] **COULD** compress to target quality
- [X] **COULD** cache processed images for future releases
- [ ] **COULD** be fast

#### backlog (februari 2022)

- [x] Planing
- [x] Basic http server
- [x] Serve images from disk
- [x] Resize image based on query parameter
- [x] Compress to target quality
- [x] Cache requested images for future requests
- [ ] decide on benchmark-method for single images
- [ ] asking for a non-exsistant id should give 404 (not 500)
- [ ] add usage log
- [ ] add cache reclaimation
- [ ] decide on how to handle handle requests for images larger than original?
- [ ] crop should keep aspect ratio

#### log (februari 2022)

- 2022-02-08: create repo and brain-dump requirements etc
- 2022-02-11: try out routing setups and settle on matryan/way routing and implement
- 2022-02-12: Add images and start defining tests for handlers
- 2022-02-12: Add ability to fetch photos
- 2022-02-13: Added a few questions that need answering
- 2022-02-13: Add parsing of query parameters
- 2022-02-13: Update documentation
- 2022-02-17: prototype resize and quality
- 2022-02-21: refactor resize and quality
- 2022-02-21: prototype cache
- 2022-02-28: have quality for gif be specifc fr gif (1-256)
- 2022-02-28: make images a package

## Requirements

- store and serve images
- preprocess images on demand
  - resize
  - compress to target quality
  - compress to target size
  - add watermark on demand
  - add instagram-like filters
- cache processed images for future releases
- be fast (needs definition)
- have configurable cache recycle rules
- have admin gui/webapp
- handle cors
- have accounts
- have account persmissions
- control access to images (possibly by cors domain?)
- have format changes
- handle vector graphics
- handle special formats (.raw etc...)
- have api for cms integration
- select focus point for crop
- be horizontaly scalable (originals storage and cache layer?)

## Questions to answer

### structure of request

- Do html and css limit query parameters?
- path or query for preprocessing
  - Query works fine
- What is easiest to work with as a front end dev?
- What parameters should be mandatory
  - none. Just image id
- What are sensible defaults when no parameters are passed?
  - thumbnail to conserve resources?
  - uncompressed original?
  - middle ground?
- Should I handle alt-texts?

### Caching

- after thinking and trying things out for a bit I am leennig towards using the filesystem as a cache. Maybe save some metadata in a datastructure of some kind.
- misses are very disruptive since we go from ~400us processing to ~400ms. About a 1000 times slower. 400ms impacts UX.
  - I will allways know the path a certain id + parameter combo will have by naming the files according approprtly.
  - on hits I will want to read from disk regardless so I might as well try to open the path
  - on miss it will not be a noticeble cost
- I decided to implement my own LRU cache. It should be thread safe but needs further verification.

## thoughts

- if used as a cdn a simple rsync could keep all cahces in sync and restore cache from master or other source on boot. Possibly even clone cache from all peers.
- set default return per image?
- should be able to use aritrary folders for images (linked by cache)
- consider using [kin-openapi](https://github.com/getkin/kin-openapi) in tests to validate endpoints agree with openapi.yaml spec

### images package API (DRAFT)

### Performance

example benchmarks run on my laptop

```bash
## 2023-05-19
goos: linux
goarch: amd64
pkg: github.com/johan-st/go-image-server
cpu: Intel(R) Core(TM) i5-10310U CPU @ 1.70GHz
Benchmark_HandleDocs-8                    354918              6096 ns/op           12613 B/op         13 allocs/op
Benchmark_HandleImg_cached-8                5510            228708 ns/op          444841 B/op         42 allocs/op
Benchmark_HandleImg_notCached-8                3         399386964 ns/op        100122525 B/op       427 allocs/op
PASS
ok      github.com/johan-st/go-image-server     16.047s

## 2023-06-01
goos: linux
goarch: amd64
pkg: github.com/johan-st/go-image-server
cpu: Intel(R) Core(TM) i5-10310U CPU @ 1.70GHz
Benchmark_HandleImg_cached-8                 939           1499931 ns/op         1386588 B/op         43 allocs/op
Benchmark_HandleImg_notCached-8                2         605024558 ns/op        159178332 B/op       306 allocs/op
PASS
ok      github.com/johan-st/go-image-server     17.842s

## 2023-06-02
goos: linux
goarch: amd64
pkg: github.com/johan-st/go-image-server
cpu: Intel(R) Core(TM) i5-10310U CPU @ 1.70GHz
Benchmark_HandleImg_cached-8                         718           1702645 ns/op         1055027 B/op         39 allocs/op
Benchmark_HandleImg_cached_concurrent-8             5318            264746 ns/op          564334 B/op         47 allocs/op
Benchmark_HandleImg_notCached-8                        1        1010868220 ns/op        318381320 B/op      2219 allocs/op
PASS
ok      github.com/johan-st/go-image-server     38.109s

## 2023-06-15 (average run)
goos: linux
goarch: amd64
pkg: github.com/johan-st/go-image-server
cpu: Intel(R) Core(TM) i5-10310U CPU @ 1.70GHz
Benchmark_HandleImg_cached-8                         607           1986089 ns/op         1241894 B/op         42 allocs/op
Benchmark_HandleImg_cached_concurrent-8             2578            458498 ns/op          657091 B/op         40 allocs/op
Benchmark_HandleImg_notCached-8                        1        1042143820 ns/op        603311696 B/op     11525 allocs/op
PASS
ok      github.com/johan-st/go-image-server     32.833s

```

## Known issues

- imagecache is not persisted between starts
- format bug: if you request a png and then a jpg of with the same parameter png will be served
- lru cache max size of 0 is not handled propperly
  
