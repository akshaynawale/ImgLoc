FROM golang:latest
# add working dir
WORKDIR /user/apps/ImgLoc
# copy code to work dir
COPY . /user/apps/ImgLoc
# change shell to bash
RUN ["/bin/bash"]
# import required libs
RUN go get github.com/golang/glog && go get github.com/rwcarlsen/goexif/exif && go get googlemaps.github.io/maps
# build app
RUN ["go", "build", "-o", "app"]
# run app
CMD ["./app", "-logtostderr", "-path", "./imgs/"]