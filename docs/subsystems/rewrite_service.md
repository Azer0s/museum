# Rewrite service (museum/service)

The rewrite service is supposed to rewrite requests. Currently, it works by replacing values in the requests and responses by a regex pattern. 

In the future, this should be compliance tested against a well known industry-standard proxy like [nginx](https://nginx.org/en/) or [Caddy](https://caddyserver.com). As of right now, the rewrite service doesn't always rewrite *correctly*.