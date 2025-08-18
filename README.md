# WatchExpense

WatchExpense is a expense reimbursement application that allows users to track their expenses and submit them for reimbursement. It is designed to be used by employees who need to keep track of their expenses and submit them for approval.

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
3. Build the application:
    ```bash
    sh scripts/build.sh
    ```
4. Run the application:
    ```bash
    ./bin/web
    ```
5. Access the application at `http://localhost:8080/health`.
6. Access the API Documentation via Swagger UI at `http://localhost:8080/docs`.

## Documentation
For detailed documentation, please refer to the [WatchExpense Documentation](https://watch-expense.onrender.com/docs) (Locally).
