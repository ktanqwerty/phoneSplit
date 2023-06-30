import requests
import json

# Set your Splitwise API access token
access_token = "YOUR_ACCESS_TOKEN_HERE"

# Set the Splitwise API endpoint URL
api_endpoint = "https://api.splitwise.com/api/v3.0/create_expense"

# Set the API request headers
headers = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {access_token}",
}

# Set the details of the expense to be added
expense_data = {
    "cost": 10.0,
    "description": "Example Expense",
    "currency_code": "USD",
    "payment": True,
    "users__0__user_id": 123456,
    "users__0__paid_share": 10.0,
    "users__0__owed_share": 0.0,
    "users__1__user_id": 789012,
    "users__1__paid_share": 0.0,
    "users__1__owed_share": 10.0,
}

# Make the API request to add the expense
response = requests.post(api_endpoint, headers=headers, data=json.dumps(expense_data))

# Check the response status code to verify the success of the request
if response.status_code == 200:
    print("Expense added successfully!")
else:
    print("Failed to add expense. Error message:", response.text)
