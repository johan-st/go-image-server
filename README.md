# go-image-server

one stop shop for serving the images you need in the formate you want.

Idea is to be able to upload a an image in high resolution and quality and have the server handle scaling etc (as per query parameters). The image server should cache images it modifies to speed up future delivery of the same request.

Idealy image workflow should be significantly simplified. Workflow could be to upload a single high-res version of each image to the image-server and let all teams fetch the sizes, qualities and formats they require for each application. Once a specific image has been requested it will be cached and instantly available on subsequent requests.

## Proof of concept

A proof of concept **MUST** be able to serve images in the requested pixel-size.
It **SHOULD** have at least two (2) quality levels and **SHOULD** be able to cache images.

## Minimum Viable Product

A MVP **MUST** be able to serve images in the requested pixel-size and **SHOULD** have rudimentary authentication/domain restrictions.

# Sprints
## [MoSCoW](https://en.wikipedia.org/wiki/MoSCoW_method)
Planing and prioritization of features and requirements to be implemented during each sprint.

Requirements not mentioned should be regarded as **WONT**

## April & May 2023
_(#backatit)_

- [x] **MUST** have tests working and passing on linux and windows
  - [x] linux 
  - [X] windows
- [x] **SHOULD** error on startup if file permissions are wrong
  - Decided to have the server try to set the permissions for image and cache folder itself if they are not permissive enough. It will only extend permissions, never reduce them.
- [ ] **SHOULD** be configurable by file or environment variables
- [ ] **COULD** have a webpage for uploading images
- [ ] **COULD** have a webpage for viewing and deleting images


### todo
- [X] Fix tests
- [X] asking for a non-exsistant id should give 404 (not 500)
  - was already fixed in previous sprint
- [X] Paths should be configurable
- [X] Refactor tests to work on the image modules API (not as part of the module)  
- [ ] decide on benchmark-method for single images
- [ ] add usage log 
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
- [ ] consider using OptFunc pattern for config. 
  - e.g: `NewServer(withPort(8080), withCacheSize(1000), withImageDefaults(imgDef), withImagePreset(imgPresetSmall), withImagePreset(imgPresetThumb))` etc...
  - https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
  - Anthony GG video: https://www.youtube.com/watch?v=MDy7JQN5MN4

### log
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

## Februari 2022


- [x] **MUST** store and serve images
- [x] **SHOULD** preprocess images on demand
  - [x] **SHOULD** resize
  - [x] **COULD** compress to target quality
- [X] **COULD** cache processed images for future releases
- [ ] **COULD** be fast

### todo

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
### log

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

# Requirements

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

# Questions to answer

## structure of request

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

## Caching
- after thinking and trying things out for a bit I am leennig towards using the filesystem as a cache. Maybe save some metadata in a datastructure of some kind.
- misses are very disruptive since we go from ~400us processing to ~400ms. About a 1000 times slower. 400ms impacts UX.
  - I will allways know the path a certain id + parameter combo will have by naming the files according approprtly.
  - on hits I will want to read from disk regardless so I might as well try to open the path
  - on miss it will not be a noticeble cost 
- I decided to implement my own LRU cache. It should be thread safe but needs further verification.

# thoughts
- if used as a cdn a simple rsync could keep all cahces in sync and restore cache from master or other source on boot. Possibly even clone cache from all peers.
- set default return per image?
- should be able to use aritrary folders for images (linked by cache)
- 
## images package API (DRAFT)


# Known issues
- imagecache is not persisted between starts
- format bug: if you request a png and then a jpg of with the same parameter png will be served