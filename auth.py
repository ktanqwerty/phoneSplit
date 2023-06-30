from google.oauth2.credentials import Credentials

creds_path = "credentialss.json"
creds = Credentials.from_authorized_user_file(creds_path, scopes=["https://www.googleapis.com/auth/gmail.readonly"])

