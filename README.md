# Todo App

A simple Todo application built with Go, MySQL, and Chi router. This app allows users to create, update, delete, and fetch todo items.

---

## Technologies Used

- **Go** (Golang) - Backend REST API
- **MySQL** - Database for storing todos
- **Chi Router** - Lightweight HTTP router for Go
- **Renderer** - For JSON responses
- **Docker** - Containerization
- **Docker Hub** - Image registry

---

## Running the Application with Docker

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed on your machine.
- MySQL database accessible at host `host.docker.internal:55000` (or adjust accordingly).
- If running locally, ensure MySQL credentials are set properly.

### Pull and Run the Docker Image

You can pull the pre-built Docker image from Docker Hub and run it directly:

```bash
docker pull disara/todo:latest
docker run -p 19000:19000 --add-host=host.docker.internal:host-gateway disara/todo:latest
