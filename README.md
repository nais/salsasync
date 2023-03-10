# salsasync

syncronize teams and users from console to salsa-storage (dependencytrack)

## Run locally

Grab .console.env and .dtrack.env file from google secret manager:

`gcloud secrets versions access latest --secret=features-dev-env --project aura-dev-d9f5 > .console.env`
`gcloud secrets versions access latest --secret=salsastorage-local --project aura-dev-d9f5 > .dtrack.env`
