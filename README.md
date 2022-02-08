# go-image-server
on stop shop for serving the images you need in the formate you want.

# Requirements Februari 2022
([MoSCoW](https://en.wikipedia.org/wiki/MoSCoW_method))
- MUST store and serve images
- SHOULD preprocess images on demand
  - SHOULD resize
  - COULD compress to target quality
  - WONT compress to target size
  - WONT add watermark on demand 
  - WONT add instagram-like filters
- COULD cache processed images for future releases
- COULD be fast 
- WONT have configurable cache recycle rules
- WONT have admin gui/webapp
- WONT be horizontaly scalable
- WONT have accounta
- WONT have account persmissions
- WONT have format changes
- WONT handle vector graphics
- WONT handle special formats (.raw etc...)
- WONT hav api for cms integration

## log
### 2022-02-08
- create repo and brain-dump requirements etc
