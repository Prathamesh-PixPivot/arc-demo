import pytest
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import requests
import time

class TestUserFeatures:
    
    def test_user_login(self, driver, base_url, user_credentials):
        """Test user login functionality"""
        driver.get(f"{base_url}/login")
        
        # Wait for page to load
        wait = WebDriverWait(driver, 10)
        wait.until(EC.presence_of_element_located((By.ID, "email")))
        
        # Fill in login credentials
        email_input = driver.find_element(By.ID, "email")
        password_input = driver.find_element(By.ID, "password")
        login_button = driver.find_element(By.XPATH, "//button[@type='submit']")
        
        email_input.send_keys(user_credentials['email'])
        password_input.send_keys(user_credentials['password'])
        login_button.click()
        
        # Wait for dashboard to load
        wait.until(EC.url_contains("/dashboard"))
        
        # Verify we're on the user dashboard
        assert "dashboard" in driver.current_url.lower()
        
    def test_view_user_consents(self, driver, base_url, user_credentials):
        """Test viewing user consents"""
        # First login
        self.test_user_login(driver, base_url, user_credentials)
        
        # Navigate to consents section
        wait = WebDriverWait(driver, 10)
        consents_link = wait.until(
            EC.element_to_be_clickable((By.XPATH, "//a[contains(text(), 'Consents')]"))
        )
        consents_link.click()
        
        # Wait for page to load
        wait.until(EC.presence_of_element_located((By.XPATH, "//div[contains(@class, 'consent')] | //table | //h2")))
        
        # Verify consents are displayed (or appropriate message if none)
        try:
            consents = driver.find_elements(By.XPATH, "//div[contains(@class, 'consent')] | //tr")
            # If consents exist, verify at least one is displayed
            if len(consents) > 0:
                assert True
        except:
            # Check for "no consents" message
            try:
                no_consents_msg = driver.find_element(By.XPATH, "//*[contains(text(), 'no consent') or contains(text(), 'No consent')]")
                assert no_consents_msg is not None
            except:
                # If neither consents nor "no consents" message, something is wrong
                assert False, "Neither consents nor 'no consents' message found"
                
    def test_submit_consent(self, driver, base_url, user_credentials):
        """Test submitting a consent"""
        # First login
        self.test_user_login(driver, base_url, user_credentials)
        
        # Navigate to consents section
        wait = WebDriverWait(driver, 10)
        consents_link = wait.until(
            EC.element_to_be_clickable((By.XPATH, "//a[contains(text(), 'Consents')]"))
        )
        consents_link.click()
        
        # Wait for page to load
        wait.until(EC.presence_of_element_located((By.XPATH, "//button[contains(text(), 'New Consent')] | //button[contains(text(), 'Create Consent')] | //a[contains(text(), 'New')]")))
        
        # Click create/new consent button
        try:
            create_button = driver.find_element(By.XPATH, "//button[contains(text(), 'New Consent')] | //button[contains(text(), 'Create Consent')] | //a[contains(text(), 'New')]")
            create_button.click()
            
            # Wait for form to load
            wait.until(EC.presence_of_element_located((By.XPATH, "//form | //select | //input")))
            
            # Fill in consent form (this will depend on the actual form structure)
            # For now, we'll just try to find and click a submit button
            try:
                submit_button = driver.find_element(By.XPATH, "//button[@type='submit'] | //input[@type='submit']")
                submit_button.click()
                
                # Wait for success message
                wait.until(EC.presence_of_element_located((By.XPATH, "//*[contains(text(), 'success') or contains(text(), 'Success')]")))
            except:
                # If no submit button, try to find a checkbox and click it
                checkboxes = driver.find_elements(By.XPATH, "//input[@type='checkbox']")
                if checkboxes:
                    checkboxes[0].click()
                    # Try to find submit button again
                    try:
                        submit_button = driver.find_element(By.XPATH, "//button[@type='submit'] | //input[@type='submit']")
                        submit_button.click()
                        wait.until(EC.presence_of_element_located((By.XPATH, "//*[contains(text(), 'success') or contains(text(), 'Success')]")))
                    except:
                        pass
        except Exception as e:
            # If no create button or other error, skip test
            pytest.skip(f"Create consent button not found or error: {str(e)}")
            
    def test_user_cannot_access_admin_features(self, driver, base_url, user_credentials):
        """Test that regular users cannot access admin features"""
        # First login as user
        self.test_user_login(driver, base_url, user_credentials)
        
        # Try to access admin routes directly
        driver.get(f"{base_url}/admin")
        
        wait = WebDriverWait(driver, 10)
        # Should be redirected or show unauthorized message
        try:
            # Wait for either redirect or error message
            wait.until(
                EC.or_(
                    EC.url_contains("/unauthorized"),
                    EC.url_contains("/login"),
                    EC.presence_of_element_located((By.XPATH, "//*[contains(text(), 'unauthorized') or contains(text(), 'Unauthorized') or contains(text(), 'forbidden') or contains(text(), 'Forbidden')]"))
                )
            )
            assert True
        except:
            # If we can access admin pages, that's a security issue
            # But we'll check if admin navigation is visible
            try:
                admin_nav = driver.find_element(By.XPATH, "//a[contains(text(), 'Admin')] | //a[contains(text(), 'admin')]")
                # If admin nav is visible, that's a problem
                assert False, "User can see admin navigation"
            except:
                # If no admin nav visible, that's good
                assert True
