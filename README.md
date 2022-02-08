# go-image-server

on stop shop for serving the images you need in the formate you want.

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
- have format changes
- handle vector graphics
- handle special formats (.raw etc...)
- hav api for cms integration
- be horizontaly scalable

## log

### 2022-02-08

- create repo and brain-dump requirements etc
