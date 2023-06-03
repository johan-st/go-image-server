# API - img.jst.dev

## /api
this page is served here

## /api/image
### GET
Lists all available image ids

## /api/image/:id
### DELETE
Deletes the image with the given id

## /api/upload
### POST
Uploads an image to the server
accept a multipart/form-data request with the image in the field "image"