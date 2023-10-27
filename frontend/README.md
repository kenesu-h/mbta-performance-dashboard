# Frontend

## Prerequisites

- [Docker Engine](https://docs.docker.com/engine/)
- [Docker Compose](https://docs.docker.com/compose/)

## Quick Setup

1. Create a .env containing the following variable:
   - VITE_BACKEND_URL: Required
2. `docker-compose up -d dev`

Note, there are some weird interactions between the host and container where you'll get this 504
dependency optimization error. If this happens to you... delete your node modules and install the
dependencies using the host instead. I realize this defeats the point of a container, but see
[here](https://github.com/vitejs/vite/discussions/8749#discussioncomment-7276723) for an
explanation.
