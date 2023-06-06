# Hacking mūsēum

By default, mūsēum will listen on port 8080. You can change this by setting the `PORT` environment variable.

mūsēum uses an ioc container to manage dependencies. You can find the container in `ioc`. The default dependency set assumes that you are not running mūsēum in a Docker container, but on your local machine (so it will use the Docker API to get the internal IP address of a container). This works fine for Linux, but on macOS you need to set both the `ENVIRONMENT` (to `development`) and `PROXY_MODE` (to `dev-proxy`) environment variables to make it work.

To get this up and running, you just need to `cd` into `dev/docker-proxy` and run `docker-compose up`. This will build and start the docker-proxy. Optionally, you can specify `DEV-PROXY-URL` to point to your dev-proxy instance (defaults to `http://localhost:3000`).

> **Note:** This is not a very performant setup. It is only meant for development purposes. This setup forwards requests to the dev proxy, which, in turn, forwards requests to the container. This means that you have two hops for every request.

The dev proxy request can also be issued manually (`http://localhost:3000/proxy/<containerId>/<port>/<path*>`).