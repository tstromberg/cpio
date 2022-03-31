# klustered-chainguard-vs-chainguard

## bighair

Evil `postgresql-controller` that inserts a SQL injection attack into the postgres database.

Deployable using: `kubectl apply -f deploy.yaml`

Notice the image name: `postgrescontroller/controller:13-alpine`

## lilhair

Added persistence for `bighair` in case they discover it too quickly. Runs `kubectl apply` every 30 seconds, masquerading as an sshd process.

Deployable using: `ssh rawkode@<worker> ./main`
