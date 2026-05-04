REGISTRY=" 172.21.22.152:8080"

podman image pull alpine:latest
podman image tag alpine:latest  172.21.22.152:8080/oci-local/artifact:latest
HTTP_PROXY=http://172.21.22.151:9090 http_proxy=http://172.21.22.151:9090 podman image push \
                  172.21.22.152:8080/oci-local/artifact:latest --tls-verify=false # --log-level=trace

# Push a Single-File Artifact

echo 'Readme Content' > readme.md


oras push --plain-http $REGISTRY/oci-local/artifact:readme \
    --artifact-type readme/example \
    ./readme.md:application/markdown


# Push a multi-file artifact

echo 'Readme Content' > readme.md
mkdir details/
echo 'Detailed Content' > details/readme-details.md
echo 'More detailed Content' > details/readme-more-details.md

oras push $REGISTRY/oci-registry/artifact:readme \
    --artifact-type readme/example\
    ./readme.md:application/markdown\
    ./details

# Discovery the manifest

oras manifest fetch --pretty $REGISTRY/oci-local/artifact:readme

# Pull an artifact

mkdir ./download

oras pull --plain-http -o ./download $REGISTRY/oci-local/artifact:readme

# View the pulled files

tree ./download


# Attach, push, and pull supply chain artifact with ORAS


# Attach a Signature to the image

echo '{"atifact": "'${IMAGE}'", "signature": "jayden hancock"}' > signature.jsonr

oras attach $IMAGE \
    --artifact-type signature/example \
    ./signature.json:application/json

oras discover -o tree $IMAGE

# Attach a sample SBOM to the image

echo '{"version": "0.0.0.0", "artifact": "'${IMAGE}'", "contents": "good"}' > sbom.json

oras attach $IMAGE \
  --rtifact-type sbom/example \
  ./sbom.json:application/jsona

# Sign and attach the SBOM

SBOM_DIGEST=$(oras discover -o json \
                --artifact-type sbom/example \
                $IMAGE | jq -r ".manifests[0].digest")

echo '{"artifact": "'$IMAGE@$SBOM_DIGEST'", "signature": "jayden hancock"}' > sbom-signature.json

oras attach $IMAGE@$SBOM_DIGEST \
  --artifact-type 'signature/example' \
  ./sbom-signature.json:application/json

# View the graph

oras discover -o tree $IMAGE


