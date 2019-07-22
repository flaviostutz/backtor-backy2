FROM golang:1.12.3 AS BUILD

RUN apt-get update && apt-get install -y libgeos-dev

RUN mkdir /backtor-restic
WORKDIR /backtor-restic

ADD go.mod .
ADD go.sum .
RUN go mod download

#now build source code
ADD . ./
RUN go build -o /go/bin/backtor-restic



FROM golang:1.12.3

    --log-level=$LOG_LEVEL \
    --conductor-url=$CONDUCTOR_API_URL \
    --repo-dir=$TARGET_DATA_PATH \
    --source-path=$SOURCE_DATA_PATH

ENV RESTIC_PASSWORD ''
ENV SOURCE_DATA_PATH '/backup-source'
ENV TARGET_DATA_PATH '/backup-repo'
ENV CONDUCTOR_API_URL ''
# ENV PRE_POST_TIMEOUT '7200'
# ENV PRE_BACKUP_COMMAND ''
# ENV POST_BACKUP_COMMAND ''

COPY --from=BUILD /go/bin/* /bin/
ADD /startup.sh /
ENTRYPOINT /startup.sh

EXPOSE 4000
