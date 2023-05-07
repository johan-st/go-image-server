# `jst_ImageServer`

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

#### parameter

| parameter       | type    | range                         | interpretation                                  |
| --------------- | ------- | ----------------------------- | ----------------------------------------------- |
| `w` / `width`   | integer | 1 or greater                  | desired width in pixels                         |
| `h` / `height`  | integer | 1 or greater                  | desired height in pixels                        |
| `f` / `format`  | string  | "jpeg" / "jpg", "png","gif"   | desired image format                            |
| `q` / `quality` | integer | 1-100 for jpeg. 1-256 for gif | jpeg: quality in percent. gif: number of colors |

#### parameters details:
- `width` / `w`: Accepts integers greater than 0. This parameter determines the width in pixels of the returned image. 
- `height` / `h`: Accepts integers greater than 0. This parameter determines the height in pixels of the returned image. 
  - If only one of width or height is specified the other will be calculated to keep the aspect ratio of the original image.
  - If both are specified the image will be cropped to the specified size. (TODO: make it crop, not stretch)
- `format` / `f`: Accepts "jpeg"/"jpg", "png" and "gif". This parameter determines the format of the returned image. 
- `quality` / `q`: quality, accepts integers. 
  - `Jpeg`: Accepts values between 1 and 100 (inclusive). Around 80 is a good value for most images.
  - `png`: Can not be compressed and will always be full quality (TODO: source)
  - `gif`: Quality is determined by the number of colors in the image. Accepts values between 1 and 256 (inclusive).



## examples

### /1/linked_image.jpeg?w=100
![linked image example](/1/linked_image.jpeg?w=200)


### /2/linked_image.jpeg?f=png&w=25&h=25
![linked image example](/2/linked_image.jpeg?f=png&w=250&h=250)

### /3?f=gif&q=4&w=700&h=100
![linked image example](/3?f=gif&q=4&w=700&h=100)
