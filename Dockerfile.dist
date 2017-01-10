FROM alpine
MAINTAINER Shijiang Wei <mountkin@gmail.com>

RUN apk update && \
  apk add --no-cache ca-certificates && \
  rm -fr /var/cache/apk

ADD bin/qingcloud-docker-network /bin/

# We deliberately don't set default values for the following environment variables.
# The user must create the container with `-e` parameter to pass in the variables.
# ENV ACCESS_KEY_ID=your-access-key-id
# ENV SECRET_KEY=your-secret-key
# ENV ZONE=sh1a

CMD ["/bin/qingcloud-docker-network"]
