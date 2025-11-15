# Blog-Go API

This is a simple blog API built with Go using the Gin web framework and MongoDB as the database. It includes user authentication with JWT, post management, and OpenAPI documentation.

## Features

*   User registration and login
*   JWT-based authentication
*   Create, read, update, and delete blog posts
*   Admin user seeding
*   OpenAPI documentation

## Technologies

*   **Go**: Programming language
*   **Gin**: Web framework
*   **MongoDB**: Database
*   **JWT**: For authentication
*   **Cloudinary**: (Optional) For media management
*   **godotenv**: For loading environment variables

## Setup

1.  **Clone the repository**:

    ```bash
    git clone <repository-url>
    cd blog-go
    ```

2.  **Environment Variables**:

    Create a `.env` file in the root directory and add the following:

    ```env

    ```

3.  **Run the application**:

    ```bash
    go mod tidy
    go run main.go
    ```

    The API will be running on `http://localhost:8000`.

## API Endpoints

*   `/register` [POST]: Register a new user
*   `/login` [POST]: Log in a user and get a JWT token
*   `/posts` [GET]: Get all posts
*   `/posts` [POST]: Create a new post (requires authentication)
*   `/posts/:id` [PUT]: Update a post (requires authentication)
*   `/posts/:id` [DELETE]: Delete a post (requires authentication)
*   `/docs` [GET]: OpenAPI documentation (HTML)
*   `/docs/openapi.yaml` [GET]: OpenAPI specification in YAML format
