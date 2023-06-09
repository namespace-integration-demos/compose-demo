nsc docker login --output_registry_to /tmp/registry
REGISTRY=`cat /tmp/registry`
nsc create --bare --cidfile /tmp/cid --duration 5m
CLUSTER_ID=`cat /tmp/cid`
nsc docker remote $CLUSTER_ID compose up --build -d
nsc expose $CLUSTER_ID --all --ingress '*=noauth' -o json > ingress.json
BACKEND=`cat ingress.json | jq -r '.[0].url'`
nsc build -t $REGISTRY/frontend --build-arg NEXT_PUBLIC_BACKEND_URL=$BACKEND --push frontend
nsc docker remote $CLUSTER_ID run -d --name frontend -p 3000:3000 $REGISTRY/frontend
nsc expose $CLUSTER_ID --container frontend
