# mūsēum 🏛
An easy to use proxy server, orchestrator and serverless runtime for your web applications

## What is mūsēum?
mūsēum (/muːˈseː.um/) is a project from the University of Vienna to provide researchers with a simple way to archive and access old web applications. Often, in the course of a research project, web applications are created to provide a user interface for data collection or analysis or simply to share ones research. These applications are mostly developed quickly and with little regard for long-term maintenance. As a result, they are often difficult to access and maintain. mūsēum provides a simple way to archive and access these applications.


<details>
<summary>
:warning: <b>This application is WIP</b> :warning:
</summary>
  
Since there is only one person working on mūsēum, progress is kinda slow (relatively speaking - the working parts might not look like much but **a lot** of work up until now has been code infrastructure). As you can see, the roadmap is still quite long so I am happy for any contribution. We plan to go stable in 2025. Maybe sooner, maybe not. You can never know with publically funded projects. 🤷

- [ ] Starting and stopping applications
  - [x] On Docker Swarm
  - [ ] On DIND
  - [ ] On K8s
- [ ] Serverless runtime
  - [ ] JS
  - [ ] WASM
- [ ] Proxy
  - [x] HTTP
  - [ ] SSE
  - [ ] WS
- [ ] Persistence
  - [ ] Resetting containers
  - [ ] Initial state
    - [ ] From NFS
    - [ ] From SMB
  - [ ] Data versioning
  - [ ] Application versioning
 - [ ] Metadata
   - [ ] OID
   - [x] Metadata sources through NATS
 - [x] Observability
   - [x] Jaeger
   - [x] Logging
- [ ] CLI tooling
  - [x] Creating exhibits
  - [ ] Deleting exhibits
  - [ ] Warming up exhibits
  - [ ] Stopping exhibits
- [ ] UI
  - [x] Loading screen
  - [ ] mūsēum UI
</details>

## How does it work?
mūsēum is fully distributed by design. Under the hood, it uses etcd to store information on running applications (which also makes mūsēum distributed). Whenever there is a request for a specific application, mūsēum will check if the application is running within the Docker Swarm. If it is, it will forward the request to the application. If it is not, it will start the application, display a loading screen and forward the request to the application once it is ready. 

Every application has to take out a "lease" on the application name. This lease is valid for a certain amount of time. If the application does not renew the lease (this is done every time mūsēum receives a request for the application), it will be removed from the Cluster. This ensures that applications that are not used for a long time will be removed from the Swarm.

## How do I use it?
mūsēum is available as a Docker image. You can find the image on Docker Hub. To run mūsēum, you need to provide the following environment variables:

* `ETCD_HOST`: The address of the etcd instance
* `ETCD_BASE_KEY`: The base key to use for etcd (optional, defaults to `museum`)
* `NATS_HOST`: The address of the NATS instance
* `NATS_BASE_KEY`: The base key to use for NATS (optional, defaults to `museum`)
* `DOCKER_HOST`: The address of the Docker Swarm (optional, defaults to `unix:///var/run/docker.sock`)
* `PROXY_MODE`: The mode to use for the proxy (optional, defaults to `swarm-ext`)
  * `swarm`: Use the Docker Swarm to start applications (assumes that mūsēum is running in a Docker Swarm)
  * `swarm-ext`: Use the Docker Swarm to start applications (assumes that mūsēum is running outside the Docker Swarm)
* `HOSTNAME`: The hostname of the mūsēum instance (optional, defaults to `localhost`)
* `PORT`: The port to listen on (optional, defaults to `8080`)
* `JAEGER_HOST`: The address of the Jaeger instance (optional)
* `ENVIRONMENT`: The environment mūsēum is running in (optional, defaults to `development`)
* `CERT_FILE`: The path to the certificate file (optional)
* `KEY_FILE`: The path to the key file (optional)
* `STARTING_TIMEOUT`: The timeout for starting an application in seconds (optional, defaults to `280`)

The proxy comes with a command line utility to manage applications. You can use it to start, stop and remove applications, etc.

As of right now, the proxy only supports Docker Swarm. We are working on adding support for Kubernetes.

### Docker Swarm compose file

```yaml
version: '3.7'
services:
  museum:
    image: museum:latest
    environment:
      ETCD_HOST: etcd
      DOCKER_HOST: unix:///var/run/docker.sock
      PROXY_MODE: swarm
      # PROXY_MODE: dind # if you want to use Docker in Docker
      HOSTNAME: museum
      PORT: 8080
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    ports:
      - "8080:8080"
    depends_on:
      - etcd
      - nats
  etcd:
    image: bitnami/etcd:latest
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
    ports:
      - "2379:2379"
  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "6222:6222"
      - "8222:8222"
```

## Exhibits

Applications are configured with so-called "exhibit" files. These files are simple YAML files that contain information on how to start the application.

```yaml
spec: v1
name: my-research-project
expose: my-webapp
lease: 1h
objects:
  - name: my-database
    image: postgres
    label: 9.6
    environment:
      POSTGRES_PASSWORD: mysecretpassword
    livecheck:
      type: tcp
      config:
        port: 5432
    mounts:
      postgres: /var/lib/postgresql/data
    volumes:
      # these volumes are always read-only, 
      # since the applications can be really old and have several security vulnerabilities,
      # we don't want to risk them being able to write (or possibly delete) any data
      - name: postgres
        driver:
          type: local
          config:
            path: /var/lib/postgresql/data
      
  - name: my-webapp
    image: my-research-project:latest
    environment:
      DATABASE_URL: postgres://postgres:mysecretpassword@my-database:5432/postgres
    livecheck:
      type: http
      config:
        path: /health
        port: 8080
        status: 200

order:
  - my-database
  - my-webapp
```

You would then start the application with `museum create my-exhibit.yml`. This will start the application and make it available at a random path (a UUID) on the proxy. You can then access the application at `http://<proxy-host>:<proxy-port>/exhibits/<path>`. You can also specify a path for the application with the `--path` flag (this is not recommended as we want to avoid path collisions).

```bash
$ museum create my-exhibit.yml
 🧑‍🎨  exhibit created successfully
 👉  http://localhost:8080/exhibit/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
```

## Accessing the applications

To access the applications, you need to know the path of the application. You can get this path by running `museum list`. 

```bash
$ museum list
───────────────────────────────────────────────────────────────────────────
🧮  my-research-project
    🔴  http://localhost:8080/exhibit/908cf715-72e8-44c7-a48d-d552b7a43918
    ⏰‎  Expired 1 hour 46 minutes 54 seconds ago
    🧺  exhibits:
        📜  db (postgres:9.6)
        📜  wordpress (my-research-project:latest)
───────────────────────────────────────────────────────────────────────────
🧮  my-other-project
    🔴  http://localhost:8080/exhibit/8122d89c-e58d-48ca-a51d-27525b1210a3
    ⏰‎  Expired 12 hours 27 minutes 43 seconds ago
    🧺  exhibits:
        📜  my-perl-app (perl:5.30)
───────────────────────────────────────────────────────────────────────────
```

### Manually stopping an application
```bash
$ museum stop my-research-project
 🛑  exhibit stopped successfully
```

### Deleting an application
```bash
$ museum delete my-research-project
 🗑  exhibit deleted successfully
```

### Renewing the lease manually
```bash
$ museum renew my-research-project 2h
 ⏲  exhibit lease renewed successfully
 👉  http://localhost:8080/exhibit/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
```

### Starting an application manually (hot start)
```bash
$ museum warmup my-research-project
 🔥  exhibit warmed up successfully
 👉  http://localhost:8080/exhibit/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
```
