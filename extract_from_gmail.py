from __future__ import print_function

from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
import base64
import email

import os.path

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError

# Set the credentials path for the Gmail API
creds_path = "credentialss.json"
SCOPES = ['https://www.googleapis.com/auth/gmail.readonly']


creds = None
    # The file token.json stores the user's access and refresh tokens, and is
    # created automatically when the authorization flow completes for the first
    # time.
if os.path.exists('token.json'):
    creds = Credentials.from_authorized_user_file('token.json', SCOPES)

if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file(
                'credentialss.json', SCOPES)
            creds = flow.run_local_server(port=0)
        # Save the credentials for the next run
        with open('token.json', 'w') as token:
            token.write(creds.to_json())

# Set the Gmail API service object
service = build('gmail', 'v1', credentials=creds)

# Set the email search query to retrieve emails from PhonePe
query = "from:noreply@phonepe.com"

# Set the maximum number of emails to retrieve
max_results = 10

# Make the API request to retrieve the emails matching the search query
try:
    response = service.users().messages().list(userId='me', q=query, maxResults=max_results).execute()
    messages = response['messages']
except HttpError as error:
    print(f'An error occurred: {error}')
    messages = []

# Loop through the retrieved messages and extract the amount and message from PhonePe emails
for message in messages:
    # print(1)
    # print(message)
    
    msg = service.users().messages().get(userId='me', id=message['id']).execute()
    payload = msg['payload']
    headers = payload['headers']
    for header in headers:
        # print(header)

        if header['name'] == 'Subject' and 'Sent' in header['value']:
            parts = payload['parts']
            for part in parts:
                if part["mimeType"] in ["text/plain"]:
                    data = base64.urlsafe_b64decode(part["body"]["data"])
                    sdata = data.decode()
                    sdata = sdata.replace(" ","")
                    msg = sdata[sdata.find('Message:')+len('Message')+1:sdata.find('Hi')]
                    amount = sdata[sdata.find('₹')+1:sdata.find('Txn')]
                    print('Message',msg)
                    print('Amount',amount)
                    allowed_chars = ['p','k','y','s']
                    if msg != "":
                        result = all(char in msg for char in allowed_chars)
                        print('result',result)

    print('-----------')
                
        
    
