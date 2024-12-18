# Exhibit files

Exhibits are described in exhibit files, a [yaml](https://yaml.org) based format. Exhibit fails have following main fields:

## spec (`string`)

Always `v1` (for now).

## name (`string`)

The name of the exhibit.

## expose (`string`)

The exhibit object to expose. The exposed object must have an exposed port.

## lease (`string`)

The lease duration as a duration string (e.g. `2h`).

## objects (`list[object]`)

The list of exhibit objects.

## rewrite (`bool`) - Optional

Determines if requests will be rewritten by the rewrite service.  Defaults to `false`. 

## order (`list[string]`) - Optional

The order in which the objects will be started. Defaults to the defined order.

## volumes (`list[volume]`) - Optional

The list of volumes used as mounts for the exhibit objects.

## meta (`list[any]`) - Optional

A list of metadata fields. This doesn't have a predefined format and will be passed on to any external application to handle.

```yaml
  phaidra-title: "Nginx example"
  pahidra-description: "A simple example of a Nginx container"
  phaidra-creator: "ariel.simulevski@univie.ac.at"
  phaidra-author-firstname: "Ariel"
  phaidra-author-lastname: "Simulevski"
  phaidra-oefos: "504017"
  phaidra-orgunit: "A495"
  phaidra-keywords:
    -
      - lang: "eng"
        value: "nginx application"
      - lang: "deu"
        value: "nginx anwendung"
```

<br>

---

<br>

# `object` 

## name (`string`)

The name of the object.

## image (`string`)

The name of the container image.

## label (`string`)

The label of the container image.

## port (`int`) - Optional

The port the exhibit object will expose a HTTP server on.

## environment (`map[string]string`) - Optional

Environment variables to use for an exhibit object. mūsēum has a simple templating engine for referencing other exhibit objects or even passing on the host name of the exhibit server.

```yaml
WORDPRESS_DB_HOST: "{{ @db }}"
WORDPRESS_WEBSITE_URL_WITHOUT_HTTP: "{{ host }}"
```

## mounts (`map[string]string`) - Optional

Maps the name of a mount to a directory in the exhibit object.

## livecheck (`livecheck`) - Optional

Defines the livecheck for an exhibit object.

<br>

---

<br>

# `livecheck`

## type (`string`)

Type of the livecheck. At the time, implemented livechecks are `http` and `exec`.

## config (`map[string]string`)

The config to use for a livecheck probe. This doesn't have a predefined format and will be passed on to the probe.

<br>

---

<br>

# `volume`

## name (`string`)

Name of the volume.

## driver (`driver`)

Config for the volume driver.

<br>

---

<br>

# `driver`

## type (`string`)

The driver type. Currently, only `local` is implemented.

## config (`map[string]string`)

The config to use for a volume driver. This doesn't have a predefined format and will be passed on to the driver.
