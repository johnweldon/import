FROM scratch

MAINTAINER John Weldon <johnweldon4@gmail.com>

COPY public /public/
ADD api api

ENV PORT 19980
ENV IMPORT_PUBLIC_DIR /public
ENV IMPORT_DB_FILE /repo.db
ENV IMPORT_VERBOSE_LOGGING=
ENV IMPORT_SAFE_IPS="127.0.0.0/8, ::1/128"

EXPOSE 19980

ENTRYPOINT ["/api"]
