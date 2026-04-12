# REQUEST → POST  /v2/oci-local/hello/blobs/uploads/
  Headers:
    User-Agent: [oras/1.3.0]
    Content-Length: [0]
    Accept-Encoding: [gzip]
  Body:

if method == POST
  and path matches "/v2/<repo>/blobs/uploads/"
  and query has no "digest"
then
    classify = StartBlobUpload


# REQUEST → PUT  /v2/oci-local/blobs/uploads/0c02d982-489f-4527-8674-c0627313a742?digest=sha256%3A67fcfcbbbd2791b8d0adbac9534169f73c5ff70972e891c3b3008a5092418ba7
  Headers:
    User-Agent: [oras/1.3.0]
    Content-Length: [47]
    Content-Type: [application/octet-stream]
    Accept-Encoding: [gzip]
  Body: <binary 47 bytes, sha256=67fcfcbbbd2791b8d0adbac9534169f73c5ff70972e891c3b3008a5092418ba7>

if method == PUT
  and path starts with "/v2/"
  and path contains "/blobs/uploads/"
  and query has key "digest"
then
    classify = CompleteBlobUpload


# REQUEST → HEAD  /v2/oci-local/hello/blobs/sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a
  Headers:
    User-Agent: [oras/1.3.0]
  Body:

if method == HEAD
  and path matches "/v2/<repo>/blobs/<digest>"
then
    classify = CheckBlobExists

# REQUEST → PUT  /v2/oci-local/blobs/uploads/0efb0122-b3f8-4003-8eef-0297b4e49743?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb
  Headers:
    Accept-Encoding: [gzip]
    Connection: [close]
    User-Agent: [containers/5.16.0 (github.com/containers/image)]
    Content-Length: [0]
    Content-Type: [application/octet-stream]
    Docker-Distribution-Api-Version: [registry/2.0]
  Body: <binary 0 bytes, sha256=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855>


if method == PUT
  and path matches "/v2/<repo>/blobs/uploads/<uuid>"
  and query contains "digest"
then
    classify = CompleteBlobUpload


# REQUEST → PATCH  /v2/oci-local/blobs/uploads/0efb0122-b3f8-4003-8eef-0297b4e49743
  Headers:
    User-Agent: [containers/5.16.0 (github.com/containers/image)]
    Content-Type: [application/octet-stream]
    Docker-Distribution-Api-Version: [registry/2.0]
    Accept-Encoding: [gzip]
    Connection: [close]
  Body: <binary 4000502 bytes, sha256=60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb>

RESPONSE ←
  Status: 202
  Latency: 12.97525ms
  Headers:
    Docker-Upload-Uuid: [0efb0122-b3f8-4003-8eef-0297b4e49743]
    Docker-Distribution-Api-Version: [registry/2.0]
    Location: [/v2/oci-local/blobs/uploads/0efb0122-b3f8-4003-8eef-0297b4e49743]
  Body:

if method == PATCH
  and path matches "/v2/<repo>/blobs/uploads/<uuid>"
then
    classify = UploadBlobChunk

# REQUEST → HEAD  /v2/oci-local/alpine/blobs/sha256:989e799e634906e94dc9a5ee2ee26fc92ad260522990f26e707861a5f52bf64e
  Headers:
    User-Agent: [containers/5.16.0 (github.com/containers/image)]
    Docker-Distribution-Api-Version: [registry/2.0]
    Connection: [close]
  Body:

RESPONSE ←
  Status: 404
  Latency: 39.598µs
  Headers:
    Docker-Distribution-Api-Version: [registry/2.0]
    Content-Type: [application/json]
  Body:

if method == HEAD
  and path matches "/v2/<repo>/blobs/<digest>"
then
    classify = CheckBlobExists
