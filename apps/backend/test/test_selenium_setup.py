import pytest
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager
from selenium.common.exceptions import WebDriverException


def test_selenium_setup():
    """Test that Selenium is properly configured"""
    driver = None
    try:
        # Setup Chrome driver with options
        options = webdriver.ChromeOptions()
        options.add_argument('--headless')  # Run in headless mode
        options.add_argument('--no-sandbox')
        options.add_argument('--disable-dev-shm-usage')
        
        # Use webdriver-manager to automatically download and manage chromedriver
        service = Service(ChromeDriverManager().install())
        driver = webdriver.Chrome(service=service, options=options)
        
        # Test that we can access a simple webpage
        driver.get("https://www.google.com")
        
        # Wait for page to load
        wait = WebDriverWait(driver, 10)
        search_box = wait.until(EC.presence_of_element_located((By.NAME, "q")))
        
        # Verify the page loaded correctly
        assert search_box is not None
        assert "Google" in driver.title
        
    except WebDriverException as e:
        # If Chrome is not installed, provide a clear error message
        if "cannot find Chrome binary" in str(e):
            pytest.skip("Chrome browser not installed. Please install Chrome to run Selenium tests.")
        else:
            # Re-raise other WebDriver exceptions
            raise
    except Exception as e:
        # Handle any other exceptions
        pytest.fail(f"Selenium setup failed: {str(e)}")
    finally:
        # Teardown
        if driver:
            driver.quit()

if __name__ == "__main__":
    try:
        test_selenium_setup()
        print("Selenium setup test passed successfully!")
    except Exception as e:
        print(f"Selenium setup test failed: {str(e)}")
        print("Please ensure Chrome is installed to run Selenium tests.")
