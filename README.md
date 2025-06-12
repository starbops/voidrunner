# voidrunner

Voidrunner is a task management application.

## Features
*(to be detailed)*

## Getting Started

### Prerequisites
- Go (version 1.21 or later recommended)
- Docker (optional, for PostgreSQL)

### Installation
1.  Clone the repository:
    ```bash
    git clone <repository-url>
    cd voidrunner
    ```
2.  Build the application:
    ```bash
    go build -o voidrunner cmd/main.go
    ```

### Running the Application
```bash
./voidrunner
```
By default, the server starts on port `:8080`.

## Storage Backends

This application supports multiple storage backends for task data:

*   **In-Memory**: Tasks are stored in memory and will be lost when the application stops. This is the default.
*   **PostgreSQL**: Tasks are stored in a PostgreSQL database, providing persistent storage.

### Configuration

The storage backend and its settings are configured via environment variables:

*   **`STORAGE_BACKEND`**: Specifies the storage backend to use.
    *   `"memory"`: (Default) Uses in-memory storage.
    *   `"postgres"`: Uses PostgreSQL database.
*   **`PG_HOST`**: Hostname of the PostgreSQL server.
    *   Defaults to `"localhost"` if `STORAGE_BACKEND` is `"postgres"`.
*   **`PG_PORT`**: Port of the PostgreSQL server.
    *   Defaults to `"5432"` if `STORAGE_BACKEND` is `"postgres"`.
*   **`PG_USER`**: Username for the PostgreSQL connection.
    *   **Required** if `STORAGE_BACKEND` is `"postgres"`.
*   **`PG_PASSWORD`**: Password for the PostgreSQL connection.
    *   **Required** if `STORAGE_BACKEND` is `"postgres"`.
*   **`PG_DBNAME`**: Database name in PostgreSQL.
    *   **Required** if `STORAGE_BACKEND` is `"postgres"`.

Example:
```bash
export STORAGE_BACKEND="postgres"
export PG_USER="myuser"
export PG_PASSWORD="mypassword"
export PG_DBNAME="mydb"
./voidrunner
```

### PostgreSQL Setup

If you choose to use the `"postgres"` storage backend, you need a running PostgreSQL server.

**Automatic Table Creation:**
The application will automatically attempt to create the necessary `tasks` table upon startup if it does not already exist in the specified database.

**Manual Table Creation (Reference):**
If you prefer to create the table manually or need to understand its structure, here is the DDL statement used by the application:
```sql
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL
);
```

## API Endpoints
*(to be detailed)*

## Contributing
*(to be detailed)*

## License
*(to be detailed)*
