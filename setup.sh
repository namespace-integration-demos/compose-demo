set -ex

nsc docker login --output_registry_to /tmp/registry
REGISTRY=`cat /tmp/registry`

#nsc create --bare --cidfile /tmp/cid --features EXP_DOCKER_CONTAINERD_STORE --experimental_from prewarm.json
nsc create --bare --cidfile /tmp/cid
CLUSTER_ID=`cat /tmp/cid`
DOCKER_CONTEXT=nsc-$CLUSTER_ID

# Create a separate docker context to ensure that we can change `buildx`'s builder.
nsc docker attach-context --name $DOCKER_CONTEXT --to $CLUSTER_ID --background

# Setup the remote builder (for caching across runs).
nsc docker buildx setup --background --create_at_startup --name nsc-remote

# Tell buildx to use the remote builder ("nsc-remote").
docker -c $DOCKER_CONTEXT buildx use nsc-remote

# Bring up the dependencies.
docker -c $DOCKER_CONTEXT compose up --build -d

# Expose the backend's ingress.
nsc expose $CLUSTER_ID --all --ingress '*=noauth' -o json > ingress.json
BACKEND=`cat ingress.json | jq -r '.[0].url'`

# Issue a build of the frontend.
docker -c $DOCKER_CONTEXT build -t $REGISTRY/frontend --build-arg NEXT_PUBLIC_BACKEND_URL=$BACKEND --push frontend
# --load is only available in buildx 0.11.0
# docker -c $DOCKER_CONTEXT build -t $REGISTRY/frontend --build-arg NEXT_PUBLIC_BACKEND_URL=$BACKEND --load frontend

# Start the frontend in the same cluster, alongside the other containers.
docker -c $DOCKER_CONTEXT run -d --name frontend -p 3000:3000 $REGISTRY/frontend

# Expose the frontend to the internet.
nsc expose $CLUSTER_ID --container frontend
