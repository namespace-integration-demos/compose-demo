set -ex

nsc docker login --output_registry_to /tmp/registry
REGISTRY=`cat /tmp/registry`

#nsc create --bare --features EXP_DOCKER_CONTAINERD_STORE --experimental_from prewarm.json
nsc create --bare -o json > cluster.json
CLUSTER_ID=`jq -r .cluster_id < cluster.json`
INGRESS_DOMAIN=`jq -r .ingress_domain < cluster.json`
DOCKER_CONTEXT=nsc-$CLUSTER_ID
BACKEND_INGRESS_NAME=backend
BACKEND_URL="${BACKEND_INGRESS_NAME}-${CLUSTER_ID}.${INGRESS_DOMAIN}"

# Create a separate docker context to ensure that we can change `buildx`'s builder.
nsc docker attach-context --name $DOCKER_CONTEXT --to $CLUSTER_ID --background

# Setup the remote builder (for caching across runs).
nsc docker buildx setup --background --create_at_startup --name nsc-remote

# Tell buildx to use the remote builder ("nsc-remote").
docker -c $DOCKER_CONTEXT buildx use nsc-remote

# Build backend and frontend in parallel
REGISTRY=$REGISTRY BACKEND_URL=$BACKEND_URL docker -c $DOCKER_CONTEXT buildx bake --push

# Bring up the dependencies.
docker -c $DOCKER_CONTEXT compose up -d

# Expose the backend's ingress.
nsc expose $CLUSTER_ID --ingress '*=noauth' --container compose-demo-backend-1 --name $BACKEND_INGRESS_NAME

# Issue a build of the frontend.
# docker -c $DOCKER_CONTEXT build -t $REGISTRY/frontend --build-arg NEXT_PUBLIC_BACKEND_URL=$BACKEND --push frontend
# --load is only available in buildx 0.11.0
# docker -c $DOCKER_CONTEXT build -t $REGISTRY/frontend --build-arg NEXT_PUBLIC_BACKEND_URL=$BACKEND --load frontend

# Start the frontend in the same cluster, alongside the other containers.
docker -c $DOCKER_CONTEXT run -d --name frontend -p 3000:3000 $REGISTRY/frontend

# Expose the frontend to the internet.
nsc expose $CLUSTER_ID --container frontend
