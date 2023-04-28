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
- [ ] Paths should be configurable
- [ ] Refactor tests to work on the image modules API (not as part of the module)  
- [ ] decide on benchmark-method for single images
- [ ] add usage log 
- [ ] add simple cache retention and reclaimation
- [ ] decide on how to handle handle requests for images larger than original?
- [ ] crop should keep aspect ratio
- [ ] handle folder creation and permissions
  - [ ] win
  - [ ] linux
- [ ] tests need to be able to run with no setup (exept conf) after clone
- [ ] app need to be able to run with no setup (except conf) after clone

### log
- 2023-04-25: look over code and tests. Plan future work
- 2023-04-26: Fix permission issues with cache and image folders. Tests now pass on linux
- 2023-04-26: slight restructure of README.md
- 2023-04-26: Fix tests (and image module) on windows
- 2023-04-26: refactor out io/util as it is deprecated
- 2023-04-28: Think through images package API (images_api_test.go)

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

# thoughts
- if used as a cdn a simple rsync could keep all cahces in sync and restore cache from master or other source on boot. Possibly even clone cache from all peers.
- set default return per image?

