#!/usr/bin/env python3

import requests
import json
import os
from dotenv import load_dotenv

load_dotenv()

API_BASE_URL = os.getenv('API_BASE_URL', 'http://localhost:8080')
ADMIN_EMAIL = os.getenv('ADMIN_EMAIL', 'admin@example.com')
ADMIN_PASSWORD = os.getenv('ADMIN_PASSWORD', 'admin123')
USER_EMAIL = os.getenv('USER_EMAIL', 'user@example.com')
USER_PASSWORD = os.getenv('USER_PASSWORD', 'user123')

ADMIN_TOKEN = None

def login_as_admin():
    """Login as admin and get authentication token"""
    global ADMIN_TOKEN
    
    login_data = {
        'email': ADMIN_EMAIL,
        'password': ADMIN_PASSWORD
    }
    
    try:
        response = requests.post(f"{API_BASE_URL}/api/v1/auth/login", json=login_data)
        if response.status_code == 200:
            ADMIN_TOKEN = response.json().get('token')
            print("Admin login successful")
            return True
        else:
            print(f"Admin login failed: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"Error during admin login: {e}")
        return False

def create_test_purpose():
    """Create a test purpose for consent forms"""
    if not ADMIN_TOKEN:
        print("Not authenticated as admin")
        return None
        
    purpose_data = {
        'name': 'Selenium Test Purpose',
        'description': 'Purpose created for Selenium testing',
        'data_objects': ['Name', 'Email', 'Phone Number']
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/fiduciary/purposes", 
            json=purpose_data,
            headers={'Authorization': f'Bearer {ADMIN_TOKEN}'}
        )
        
        if response.status_code == 201:
            purpose = response.json()
            print(f"Created test purpose: {purpose.get('name')}")
            return purpose
        elif response.status_code == 409:
            print("Test purpose already exists")
            # Try to get existing purpose
            response = requests.get(
                f"{API_BASE_URL}/api/v1/fiduciary/purposes", 
                headers={'Authorization': f'Bearer {ADMIN_TOKEN}'}
            )
            if response.status_code == 200:
                purposes = response.json()
                for purpose in purposes:
                    if purpose.get('name') == 'Selenium Test Purpose':
                        return purpose
            return None
        else:
            print(f"Failed to create test purpose: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Error creating test purpose: {e}")
        return None

def create_test_consent_form(purpose_id):
    """Create a test consent form"""
    if not ADMIN_TOKEN:
        print("Not authenticated as admin")
        return None
        
    form_data = {
        'title': 'Selenium Test Consent Form',
        'description': 'Consent form created for Selenium testing',
        'terms_url': 'https://example.com/terms',
        'privacy_url': 'https://example.com/privacy',
        'purposes': [purpose_id],
        'dataObjects': {
            'Name': 'User full name',
            'Email': 'User email address',
            'Phone Number': 'User phone number'
        },
        'vendors': {
            'analytics': ['Google Analytics'],
            'marketing': ['Facebook Ads']
        }
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/fiduciary/consent-forms", 
            json=form_data,
            headers={'Authorization': f'Bearer {ADMIN_TOKEN}'}
        )
        
        if response.status_code == 201:
            form = response.json()
            print(f"Created test consent form: {form.get('title')}")
            return form
        elif response.status_code == 409:
            print("Test consent form already exists")
            # Try to get existing form
            response = requests.get(
                f"{API_BASE_URL}/api/v1/fiduciary/consent-forms", 
                headers={'Authorization': f'Bearer {ADMIN_TOKEN}'}
            )
            if response.status_code == 200:
                forms = response.json()
                for form in forms:
                    if form.get('title') == 'Selenium Test Consent Form':
                        return form
            return None
        else:
            print(f"Failed to create test consent form: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Error creating test consent form: {e}")
        return None

def create_test_api_key():
    """Create a test API key"""
    if not ADMIN_TOKEN:
        print("Not authenticated as admin")
        return None
        
    api_key_data = {
        'name': 'Selenium Test API Key',
        'scopes': ['public_forms:read', 'consents:write']
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/fiduciary/api-keys", 
            json=api_key_data,
            headers={'Authorization': f'Bearer {ADMIN_TOKEN}'}
        )
        
        if response.status_code == 201:
            api_key = response.json()
            print(f"Created test API key: {api_key.get('name')}")
            return api_key
        elif response.status_code == 409:
            print("Test API key already exists")
            return None
        else:
            print(f"Failed to create test API key: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Error creating test API key: {e}")
        return None

def setup_all_test_data():
    """Set up all test data"""
    print("Setting up test data...")
    
    # Login as admin
    if not login_as_admin():
        return False
    
    # Create test purpose
    purpose = create_test_purpose()
    if not purpose:
        return False
    
    # Create test consent form
    form = create_test_consent_form(purpose['id'])
    if not form:
        return False
    
    # Create test API key
    api_key = create_test_api_key()
    if not api_key:
        return False
    
    print("All test data setup successfully!")
    return True

if __name__ == "__main__":
    success = setup_all_test_data()
    if success:
        print("\nTest data is ready for Selenium testing.")
        print("You can now run the Selenium tests.")
    else:
        print("\nFailed to set up test data.")
        print("Please check the server is running and credentials are correct.")
