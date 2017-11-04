FROM scratch
MAINTAINER John Weldon <johnweldon4@gmail.com>
ADD import import
ENTRYPOINT ["/import"]
