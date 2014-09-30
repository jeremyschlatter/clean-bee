FROM scratch
MAINTAINER Jeremy Schlatter <jeremy.schlatter@gmail.com>
ADD . /
EXPOSE 80
ENTRYPOINT ["/clean-bee"]
