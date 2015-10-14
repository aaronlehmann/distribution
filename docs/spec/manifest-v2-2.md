# Image Manifest Version 2, Schema 2

This document outlines the format of of the V2 image manifest, schema version 2.
The original (and provisional) image manifest for V2 (schema 1), was introduced
in the Docker daemon in the [v1.3.0
release](https://github.com/docker/docker/commit/9f482a66ab37ec396ac61ed0c00d59122ac07453)
and is specified in the [schema 1 manifest definition](./manifest-v2-1.md)

This second schema version has two primary goals. The first is to allow
multi-architecture images, through a "fat manifest" which references image
manifests for platform-specific versions of an image. The second is to
move the Docker engine towards content-addressable images, by supporting
an image model where the image's configuration can be hashed to generate
an ID for the image.

# Media Types

The following media types are used by the manifest formats described here, and
the resources they reference:

- `application/vnd.docker.distribution.manifest.v1+json`: schema1 (existing manifest format)
- `application/vnd.docker.distribution.manifest.v2+json`: New image manifest format (schemaVersion = 2)
- `application/vnd.docker.distribution.manifest.list.v2+json`: Manifest list, aka "fat manifest"
- `application/vnd.docker.image.rootfs.diff.tar.gzip`: "Layer", as a gzipped tar
- `application/vnd.docker.container.image.v1+json`: Container config JSON

## Manifest List

The manifest list is the "fat manifest" which points to specific image manifests
for one or more platforms. Its use is optional, and relatively few images will
use one of these manifests. A client will distinguish a manifest list from an
image manifest based on the Content-Type returned in the HTTP response.

## *Manifest List* Field Descriptions

- **`schemaVersion`** *int*
	
  This field specifies the image manifest schema version as an integer. A
  manifest list uses version `3` to distinguish it from an image manifest.

- **`manifests`** *array*

    The manifests field contains a list of manifests for specific platforms.

    Fields of a object in the manifests list are:
    
    - **`mediaType`** *string*
    
        The MIME type of the referenced object. This will generally be
        `application/vnd.docker.image.manifest.v2+json`, but it could also
        be `application/vnd.docker.image.manifest.v1+json` if the manifest
        list references a legacy schema-1 manifest.
    
    - **`size`** *int*
    
        The size in bytes of the object. This field exists so that a client
        will have an expected size for the content before validating. If the
        length of the retrieved content does not match the specified length,
        the content should not be trusted.
    
    - **`digest`** *string*

        The digest of the content, as defined by the
        [Registry V2 HTTP API Specificiation](https://docs.docker.com/registry/spec/api/#digest-parameter).

    - **`labels`** *object*

        Labels may be keyed to *any* JSON object, allowing content creators to
        annotate their manifest with any additional metadata beyond those already
        defined by this manifest format or contained within the target object. 

## Example Manifest List

*Example showing a simple manifest list pointing to image manifests for two platforms:*
```json
{
  "schemaVersion": 3,
  "manifests": [
    {
      "mediaType": "application/vnd.docker.image.manifest.v2+json",
      "size": 7143,
      "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
      "labels": {
        "os": "linux",
        "arch": "ppc64le"
      },
    },
    {
      "mediaType": "application/vnd.docker.image.manifest.v2+json",
      "size": 7682,
      "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
      "labels": {
        "os": "linux",
        "arch": "x86-64"
      }
    }
  ]
}
```

# Image Manifest

The image manifest provides a configuration and a set of layers for a container
image. It's the direct replacement for the schema-1 manifest.

## *Image Manifest* Field Descriptions

- **`schemaVersion`** *int*
	
  This field specifies the image manifest schema version as an integer. An
  image manifest uses version `2`.

- **`config`** *object*

    The config field references a configuration object for a container, by
    digest. This configuration item is a JSON blob that the runtime uses
    to set up the container. This new schema uses a tweaked version
    of this configuration to allow image content-addressability on the
    daemon side.

    Fields of a config object are:
    
    - **`mediaType`** *string*
    
        The MIME type of the referenced object. This should generall be
        `application/vnd.docker.container.image.v1+json`.
    
    - **`size`** *int*
    
        The size in bytes of the object. This field exists so that a client
        will have an expected size for the content before validating. If the
        length of the retrieved content does not match the specified length,
        the content should not be trusted.
    
    - **`digest`** *string*

        The digest of the content, as defined by the
        [Registry V2 HTTP API Specificiation](https://docs.docker.com/registry/spec/api/#digest-parameter).

    - **`labels`** *object*

        Labels may be keyed to *any* JSON object, allowing content creators to
        annotate their manifest with any additional metadata beyond those already
        defined by this manifest format or contained within the target object.

- **`layers`** *array*

    The layer list is ordered starting from the base image (opposite order of schema1).
    The `baseLayer` label is intended to improve human comprehensibility of the layer
    list, but it will be ignored by the Docker engine.

    Fields of an item in the layers list are:
    
    - **`mediaType`** *string*
    
        The MIME type of the referenced object. This should
        generally be `application/vnd.docker.image.rootfs.diff.tar.gzip`.
    
    - **`size`** *int*
    
        The size in bytes of the object. This field exists so that a client
        will have an expected size for the content before validating. If the
        length of the retrieved content does not match the specified length,
        the content should not be trusted.
    
    - **`digest`** *string*

        The digest of the content, as defined by the
        [Registry V2 HTTP API Specificiation](https://docs.docker.com/registry/spec/api/#digest-parameter).

    - **`labels`** *object*

        Labels may be keyed to *any* JSON object, allowing content creators to
        annotate their manifest with any additional metadata beyond those already
        defined by this manifest format or contained within the target object. 

## Example Image Manifest

*Example showing an image manifest:*
```json
{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.docker.container.image.v1+json",
        "size": 7023,
        "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7",
        "labels": {
            "createdAt": "2015-09-16T16:29:07.952493971-07:00",
            "version": "3.1.4-a159+265"
        }
    },
    "layers": [
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 32654,
            "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
            "labels": {
              "baseLayer": true
            }
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 16724,
            "digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 73109,
            "digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
        }
    ],
}
```
