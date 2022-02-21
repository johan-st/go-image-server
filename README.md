# go-image-server

on stop shop for serving the images you need in the formate you want.

Idea is to be able to upload a an image in high resolution and quality and have the server handle scaling etc (as per query parameters). The image server should cache images it modifies to speed up future delivery of the same request.

Idealy image workflow should be significantly simplified. Workflow could be to upload a single high-res version of each image to the image-server and let all teams fetch the sizes, qualities and formats they require for each application. Once a specific image has been requested it will be cached and instantly available on subsequent requests.

## Proof of concept

A proof of concept **MUST** be able to serve images in the requested pixel-size.
It **SHOULD** have at least two (2) quality levels and **SHOULD** be able to cache images.

## Minimum Viable Product

A MVP **MUST** be able to serve images in the requested pixel-size and **SHOULD** have rudimentary authentication/domain restrictions.

## Februari 2022

### [MoSCoW](https://en.wikipedia.org/wiki/MoSCoW_method)

Requirements not mentioned should be regarded as **WONT**

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
- [ ] add usage log 
- [ ] add cache reclaimation

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

## log

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

# Questions to answer

## structure of request

- Do html and css limit query parameters?
- path or query for preprocessing
  - Query works fine
- What is easiest to work with as a front end dev?
- What parameters should be mandatory
  - none. Just image id
- Should I have presets for quality?
- What are sensible defaults when no parameters are passed?
  - thumbnail to conserve resources?
  - uncompressed original?
- Should I handle alt-texts?

# thoughts
- if used as a cdn a simple rsync could keep all cahces in sync and restore cache from master or other source on boot. Possibly even clone cache from all peers.