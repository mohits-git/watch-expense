# WatchExpense

WatchExpense is a expense reimbursement application that allows users to track their expenses and submit them for reimbursement. It is designed to be used by employees who need to keep track of their expenses and submit them for approval.

## Documentation
For detailed documentation, please refer to the [WatchExpense Documentation](https://watch-expense.onrender.com/docs).

## Setup
To set up the WatchExpense application, follow these steps:
1. Clone the repository:
   ```bash
   git clone https://github.com/mohits-git/watch-expense
    cd watch-expense
    ```
2. Install the dependencies:
    ```bash
    go mod tidy
    ```
3. Set up the swagger ui for documentation:
    ```bash
    sh scripts/setup-swagger.sh
    ```
    **Note:** Ensure you have `unzip` installed on your system to extract the Swagger UI files.
4. Build the application:
    ```bash
    go build -o watch-expense
    ```
5. Run the application:
    ```bash
    ./watch-expense
    ```
6. Access the application at `http://localhost:8080/health`.
7. Access the API Documentation via Swagger UI at `http://localhost:8080/docs`.
