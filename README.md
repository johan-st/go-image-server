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

- **MUST** store and serve images
- **SHOULD** preprocess images on demand
  - **SHOULD** resize
  - **COULD** compress to target quality
- **COULD** cache processed images for future releases
- **COULD** be fast

### todo

- [x] Planing
- [ ] Basic http server
- [ ] Serve images from disk
- [ ] Resize image based on query parameter
- [ ] Compress to target quality
- [ ] Cache requested images for future requests
- [ ] decide on benchmark-method for single images

# Requirements

- store and serve images
- preprocess images on demand
  - resize
  - compress to target quality
  - compress to target size
  - add watermark on demand
  - add instagram-like filters
- cache processed images for future releases
- be fast
- have configurable cache recycle rules
- have admin gui/webapp
- handle cors
- have accounts
- have account persmissions
- control access to images (possibly by cors domain?)
- have format changes
- handle vector graphics
- handle special formats (.raw etc...)
- hav api for cms integration
- select focus point for crop
- be horizontaly scalable (originals storage and cache layer?)

## log

- 2022-02-08: create repo and brain-dump requirements etc
- 2022-02-11: try out routing setups and settle on matryan/way routing and implement
- 2022-02-12: Add images and start defining tests for handlers
