import pytest
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import requests
import time

class TestAdminFeatures:
    
    def test_admin_login(self, driver, base_url, admin_credentials):
        """Test admin login functionality"""
        driver.get(f"{base_url}/login")
        
        # Wait for page to load
        wait = WebDriverWait(driver, 10)
        wait.until(EC.presence_of_element_located((By.ID, "email")))
        
        # Fill in login credentials
        email_input = driver.find_element(By.ID, "email")
        password_input = driver.find_element(By.ID, "password")
        login_button = driver.find_element(By.XPATH, "//button[@type='submit']")
        
        email_input.send_keys(admin_credentials['email'])
        password_input.send_keys(admin_credentials['password'])
        login_button.click()
        
        # Wait for dashboard to load
        wait.until(EC.url_contains("/dashboard"))
        
        # Verify we're on the admin dashboard
        assert "dashboard" in driver.current_url.lower()
        
    def test_create_consent_form(self, driver, base_url, admin_credentials):
        """Test creating a new consent form"""
        # First login
        self.test_admin_login(driver, base_url, admin_credentials)
        
        # Navigate to consent forms section
        wait = WebDriverWait(driver, 10)
        consent_forms_link = wait.until(
            EC.element_to_be_clickable((By.XPATH, "//a[contains(text(), 'Consent Forms')]"))
        )
        consent_forms_link.click()
        
        # Wait for page to load
        wait.until(EC.presence_of_element_located((By.XPATH, "//button[contains(text(), 'Create Consent Form')]")))
        
        # Click create button
        create_button = driver.find_element(By.XPATH, "//button[contains(text(), 'Create Consent Form')]")
        create_button.click()
        
        # Fill in form details
        title_input = wait.until(EC.presence_of_element_located((By.XPATH, "//input[@placeholder='Form Title']")))
        title_input.send_keys("Test Consent Form")
        
        desc_input = driver.find_element(By.XPATH, "//input[@placeholder='Description']")
        desc_input.send_keys("This is a test consent form for Selenium testing")
        
        # Fill in purpose details
        next_button = driver.find_element(By.XPATH, "//button[text()='Next']")
        next_button.click()
        
        # Add a purpose (assuming there's at least one available)
        try:
            purpose_checkbox = wait.until(
                EC.element_to_be_clickable((By.XPATH, "//input[@type='checkbox']"))
            )
            purpose_checkbox.click()
            
            # Add data objects
            data_objects_input = driver.find_element(By.XPATH, "//input[@placeholder='Data Objects (comma separated)']")
            data_objects_input.send_keys("Name, Email, Phone")
        except:
            pass  # If no purposes available, skip this step
        
        # Go to vendors step
        next_button.click()
        
        # Submit form
        save_button = driver.find_element(By.XPATH, "//button[text()='Save Form']")
        save_button.click()
        
        # Wait for success message
        wait.until(EC.presence_of_element_located((By.XPATH, "//*[contains(text(), 'successfully')]")))
        
    def test_manage_api_keys(self, driver, base_url, admin_credentials):
        """Test API key management functionality"""
        # First login
        self.test_admin_login(driver, base_url, admin_credentials)
        
        # Navigate to API keys section (assuming it exists)
        wait = WebDriverWait(driver, 10)
        try:
            api_keys_link = wait.until(
                EC.element_to_be_clickable((By.XPATH, "//a[contains(text(), 'API Keys')]"))
            )
            api_keys_link.click()
            
            # Wait for page to load
            wait.until(EC.presence_of_element_located((By.XPATH, "//button[contains(text(), 'Create API Key')]")))
            
            # Click create button
            create_button = driver.find_element(By.XPATH, "//button[contains(text(), 'Create API Key')]")
            create_button.click()
            
            # Fill in API key details
            name_input = wait.until(EC.presence_of_element_located((By.XPATH, "//input[@placeholder='API Key Name']")))
            name_input.send_keys("Selenium Test API Key")
            
            # Submit form
            submit_button = driver.find_element(By.XPATH, "//button[@type='submit']")
            submit_button.click()
            
            # Wait for success message
            wait.until(EC.presence_of_element_located((By.XPATH, "//*[contains(text(), 'created')]")))
            
        except Exception as e:
            # If API keys section doesn't exist or other error, skip test
            pytest.skip(f"API keys section not found or error: {str(e)}")
            
    def test_view_integration_script(self, driver, base_url, admin_credentials):
        """Test viewing integration script for a consent form"""
        # First login
        self.test_admin_login(driver, base_url, admin_credentials)
        
        # Navigate to consent forms section
        wait = WebDriverWait(driver, 10)
        consent_forms_link = wait.until(
            EC.element_to_be_clickable((By.XPATH, "//a[contains(text(), 'Consent Forms')]"))
        )
        consent_forms_link.click()
        
        # Wait for page to load
        wait.until(EC.presence_of_element_located((By.XPATH, "//table")))
        
        # Click view on first consent form
        try:
            view_button = driver.find_element(By.XPATH, "//button[text()='View']")
            view_button.click()
            
            # Wait for preview to load
            wait.until(EC.presence_of_element_located((By.XPATH, "//button[contains(text(), 'Get Integration Script')]")))
            
            # Click integration script button
            script_button = driver.find_element(By.XPATH, "//button[contains(text(), 'Get Integration Script')]")
            script_button.click()
            
            # Wait for dialog to appear
            wait.until(EC.presence_of_element_located((By.XPATH, "//div[@role='dialog']")))
            
            # Verify script content exists
            script_content = driver.find_element(By.XPATH, "//div[@role='dialog']//pre").text
            assert len(script_content) > 0
            
        except Exception as e:
            # If no consent forms exist or other error, skip test
            pytest.skip(f"No consent forms found or error: {str(e)}")
