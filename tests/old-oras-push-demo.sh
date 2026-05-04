#!/bin/bash

# Create a named stub
cat <<EOF > ./myservice.json
{
  "name": "io.wabbit-networks.netmonitor:2022-Q1-M1",
  "documentation": "https://www.wabbit-networks.io"
}
EOF

# Push a named reference to the registry
oras push 172.21.22.152:8080/oci-local/net-monitor:2022-Q1-M1 \
	--manifest-config /dev/null:application/json \ 
	./myservice.json json:application/json

# Create an SBOM for the service
cat <<EOF > ./sbom.json
{
  "version": "1.0.0.0",
  "id": "wabbit-networks/services/net-monitor:2202-abc123",
  "contents": [
    { "name": "value" },
    { "name": "value" }
  ]
}
EOF

# Push the SBOM for the service
oras push 172.21.22.152:8080/oci-local/net-monitor \
	--artifact-type 'sbom/example' \
	--subject 172.21.22.152:8080/oci-local/net-monitor:2022-Q1-M1 \
	./sbom.json:application/json

# Discovery graph
oras discover -o tree 172.21.22.152:8080/oci-local/net-monitor:2022-Q1-M1

# Create a claim
cat <<EOF > ./claims.json
{
  "claim-created": "2022-04-20T08:53:09.42",
  "claim-identity": "<identifier>",
  "subject": "wabbit-networks/service/net-monitor:build-abc123",
  "claims": [
    "gov.nist.csrc.ssdf.1.1": "true"
  ]
}
EOF

# Push the claim
oras push 172.21.22.152:8080/oci-local/net-monitor \ 
	--artifact-type 'claims/example' \
	--subject 172.21.22.152:8080/oci-local/net-monitor:2022-Q1-M1 \
	./claims.json:application/json

# Push just some annotations
cat <<EOF > ./annotations.json
{
  "\$manifest": {
    "io.acme-rockets.policy.scanned": "policy12",
    "io.cncf.oras.arifact.eol": "2022-05-31",
  }
}
EOF

cat ./annotations.json|jq


# Push the annotations

# Discovery the graph

# View the claims

# Copy the graph
