# jst_ImageServer

a go image server

## NOTICE!

This server is still under development. Incomplete features and issues are to be expected.

## routes

### GET /

This documentation resides here

### GET /:image_id

Requests to any path beyond root ("/") are treated as an image request.
This path returns the original image by id

### GET /:image_id/:desired_filename.jpeg

Subsequent path does not change the response but is helpfull for naming the file fetched.
The titular example return a file named desired_filname.jpg

## Preprocessing

### How to get the size you want?

Describing the image you want is done through query parameters added to the url.

#### parameters are:
- q: quality, accepts integers between 1 and 100 (inclusive). This parameter determines the rate of compression from none (100% quality) to destructive (1% quality).
- w: width, accepts integers greater than 0. This parameter determines the width in pixels of te returned image.
- h: height, accepts integers greater than 0. This parameter determines the height in pixels of te returned image.


#### Quality have different meanings depending on format. 

| format                | range | enterpretation                   |
| --------------------- | ----- | -------------------------------- |
| jpeg (default format) | 1-100 | % quality                        |
| gif                   | 1-256 | nr of colors                     |
| png                   |       | is ignored.  always full quality |