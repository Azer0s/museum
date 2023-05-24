# mūsēum 🏛
The fast, easy to use proxy server for your old web applications

## What is mūsēum?
mūsēum (/muːˈseː.um/) is a project from the University of Vienna to provide researchers with a simple way to archive and access old web applications. Often, in the course of a research project, web applications are created to provide a user interface for data collection or analysis or simply to share ones research. These applications are often developed quickly and with little regard for long-term maintenance. As a result, they are often difficult to access and maintain. mūsēum provides a simple way to archive and access these applications.

## How does it work?
mūsēum is fully distributed by design. Under the hood, it uses etcd to store information on running applications (which also makes mūsēum distributed). Whenever there is a request for a specific application, mūsēum will check if the application is running within the Docker Swarm. If it is, it will forward the request to the application. If it is not, it will start the application, display a loading screen and forward the request to the application once it is ready. 

Every application has to take out a "lease" on the application name. This lease is valid for a certain amount of time. If the application does not renew the lease (this is done every time mūsēum receives a request for the application), it will be removed from the Cluster. This ensures that applications that are not used for a long time will be removed from the Swarm.

## How do I use it?
mūsēum is available as a Docker image. You can find the image on Docker Hub. To run mūsēum, you need to provide the following environment variables:

* `ETCD_HOST`: The address of the etcd instance
* `ETCD_BASE_KEY`: The base key to use for etcd (optional, defaults to `museum`)
* `DOCKER_HOST`: The address of the Docker Swarm (optional, defaults to `unix:///var/run/docker.sock`)
* `HOSTNAME`: The hostname of the mūsēum instance (optional, defaults to `localhost`)
* `PORT`: The port to listen on (optional, defaults to `8080`)
* `JAEGER_HOST`: The address of the Jaeger instance (optional)
* `ENVIRONMENT`: The environment mūsēum is running in (optional, defaults to `development`)

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
  etcd:
    image: bitnami/etcd:latest
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
    ports:
      - 2379:2379
```

## Exhibits

Applications are configured with so-called "exhibit" files. These files are simple YAML files that contain information on how to start the application.

```yaml
name: my-research-project
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
 👉  http://localhost:8080/exhibits/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
```

## Accessing the applications

To access the applications, you need to know the path of the application. You can get this path by running `museum list`. 

```bash
$ museum list
> my-research-project
    👉 http://localhost:8080/exhibits/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
    ⏲ Expires in 23 minutes and 59 seconds from now
    📦 exhibits:
        📦 my-database (postgres:9.6)
        📦 my-webapp (my-research-project:latest)
> my-other-project
    👉 http://localhost:8080/exhibits/3b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
    ⏲ Expires in 1 hour, 11 minutes and 16 seconds from now
    📦 exhibits:
        📦 my-perl-app (perl:5.30)
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
 👉  http://localhost:8080/exhibits/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
```

### Starting an application manually (hot start)
```bash
$ museum warmup my-research-project
 🔥  exhibit warmed up successfully
 👉  http://localhost:8080/exhibits/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
```
