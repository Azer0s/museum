# mūsēum
The fast, easy to use proxy server for your old web applications

## What is mūsēum?
mūsēum (/muːˈseː.um/) is a project from the University of Vienna to provide researchers with a simple way to archive and access old web applications. Often, in the course of a research project, web applications are created to provide a user interface for data collection or analysis or simply to share ones research. These applications are often developed quickly and with little regard for long-term maintenance. As a result, they are often difficult to access and maintain. mūsēum provides a simple way to archive and access these applications.

## How does it work?
mūsēum is fully distributed by design. Under the hood, it uses Redis to store information on running applications and Kafka to communicate between different mūsēum instances. Whenever there is a request for a specific application, mūsēum will check if the application is running within the Docker Swarm. If it is, it will forward the request to the application. If it is not, it will start the application, display a loading screen and forward the request to the application once it is ready. 

Every application has to take out a "lease" on the application name. This lease is valid for a certain amount of time. If the application does not renew the lease (this is done every time mūsēum receives a request for the application), it will be removed from the Cluster. This ensures that applications that are not used for a long time will be removed from the Swarm.

## How do I use it?
mūsēum is available as a Docker image. You can find the image on Docker Hub. To run mūsēum, you need to provide the following environment variables:

* `KAFKA_BROKERS`: A comma-separated list of Kafka brokers
* `KAFKA_TOPIC`: The Kafka topic to use
* `REDIS_HOST`: The hostname of the Redis instance
* `REDIS_PORT`: The port of the Redis instance
* `DOCKER_HOST`: The hostname of the Docker Swarm
* `DOCKER_PORT`: The port of the Docker Swarm
* `HOSTNAME`: The hostname of the mūsēum instance

The proxy comes with a command line utility to manage applications. You can use it to start, stop and remove applications. You can also use it to list all running applications.

Applications are configured with so-called "exhibit" files. These files are simple YAML files that contain information on how to start the application.

```yaml
name: my-research-project
exhibits:
  - name: my-database
    image: postgres:9.6
    environment:
      POSTGRES_PASSWORD: mysecretpassword
    livecheck:
      type: tcp
      port: 5432
    mounts:
      - postgres:/var/lib/postgresql/data
    volumes:
      # these volumes are always read-only, 
      # since the applications can be really old and have several security vulnerabilities,
      # we don't want to risk them being able to write (or possibly delete) any data
      - name: postgres
        driver:
          type: local
          path: /var/lib/postgresql/data
      
  - name: my-webapp
    image: my-research-project:latest
    environment:
      DATABASE_URL: postgres://postgres:mysecretpassword@my-database:5432/postgres
    livecheck:
      type: http
      path: /health
      port: 8080
      status: 200

order:
  - my-database
  - my-webapp
```

You would then start the application with `museum create my-exhibit.yml`. This will start the application and make it available at a random path (a UUID) on the proxy. You can then access the application at `http://<proxy-host>:<proxy-port>/<path>`. You can also specify a path for the application with the `--path` flag (this is not recommended as we want to avoid path collisions).

## Accessing the applications

To access the applications, you need to know the path of the application. You can get this path by running `museum list`. 

```bash
$ museum list
> my-research-project
    🚗 Path: http://localhost:8080/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
    ⏲ Expires: 23 minutes and 59 seconds from now
    🔒 HTTPS: false 🔴
    📦 exhibits:
        📦 my-database (postgres:9.6)
        📦 my-webapp (my-research-project:latest)
> my-other-project
    🚗 Path: http://localhost:8080/3b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
    ⏲ Expires: 1 hour, 11 minutes and 16 seconds from now
    🔒 HTTPS: true 🟢
    📦 exhibits:
        📦 my-perl-app (perl:5.30)
```

### Renewing the lease manually
```bash
$ museum renew my-research-project 2h
> my-research-project
    🚗 Path: http://localhost:8080/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
    ⏲ Expires: 2 hours, 0 minutes and 0 seconds from now
    🔒 HTTPS: false 🔴
    📦 exhibits:
        📦 my-database (postgres:9.6)
        📦 my-webapp (my-research-project:latest)
```

### Starting an application manually (hot start)
```bash
$ museum warmup my-research-project
> my-research-project
    🚗 Path: http://localhost:8080/5b3c0e3e-1b5a-4b1f-9b1f-1b5a4b1f9b1f
    ⏲ Expires: 2 hours, 0 minutes and 0 seconds from now
    🔒 HTTPS: false 🔴
    📦 exhibits:
        📦 my-database (postgres:9.6)
        📦 my-webapp (my-research-project:latest)
```
