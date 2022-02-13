# go-image-server

## routes

### GET /

This documentation resides here

### GET /:image_id/:desired_filename.jpeg

Requests to any path beyond root ("/") are treated as an image request.
Subsequent path does not change the response but is helpfull for naming the "file" fetched.
The titular example return a file name "desired_filname.jpg"
