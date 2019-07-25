# backtor-restic

Backtor Conductor worker for performing Restic backups

## Usage

* Create a docker-compose.yml:

```yml
version: '3.5'

services:

  backtor-restic:
    image: flaviostutz/backtor-restic
    environment:
      - RESTIC_PASSWORD=123
      - LOG_LEVEL=debug
      - SOURCE_DATA_PATH=/backup-source/
      - REPO_DIR=/backup-repo/
      - CONDUCTOR_API_URL=http://backtor-conductor:8080/api

  backtor:
    image: flaviostutz/backtor
    restart: always
    ports:
      - 6000:6000
    environment:
      - LOG_LEVEL=debug
      - CONDUCTOR_API_URL=http://backtor-conductor:8080/api

  backtor-conductor:
    image: flaviostutz/backtor-conductor
    restart: always
    ports:
      - 8080:8080
    environment:
      - DYNOMITE_HOSTS=dynomite:8102:us-east-1c
      - ELASTICSEARCH_URL=elasticsearch:9300
      - LOADSAMPLE=false
      - PROVISIONING_UPDATE_EXISTING_TASKS=false

  dynomite:
    image: flaviostutz/dynomite:0.7.5
    restart: always
    ports:
      - 8102:8102

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:5.6.8
    restart: always
    environment:
      - "ES_JAVA_OPTS=-Xms512m -Xmx1000m"
      - transport.host=0.0.0.0
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports:
      - 9200:9200
      - 9300:9300
    logging:
      driver: "json-file"
      options:
        max-size: "20MB"
        max-file: "5"

  conductor-ui:
    image: flaviostutz/conductor-ui
    restart: always
    environment:
      - WF_SERVER=http://backtor-conductor:8080/api/
    ports:
      - 5000:5000

```

* Run 'docker-compose up'

* See logs for seeing worker to run tasks

* See Conductor UI at http://localhost:5000 to check for tasks being COMPLETED

