# Microservice Analytics

A simple analytics microservice built with Go Fiber and Swagger.

## Running the Application

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Generate Swagger docs:
   ```bash
   go run github.com/swaggo/swag/cmd/swag@latest init
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

4. Access the API:
   - API: http://localhost:3000/
   - Swagger UI: http://localhost:3000/swagger/index.html