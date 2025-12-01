import pytest
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import WebDriverException
from webdriver_manager.chrome import ChromeDriverManager
from dotenv import load_dotenv
import os

# Load environment variables
load_dotenv()

@pytest.fixture(scope="session")
def driver():
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
        driver.implicitly_wait(10)
        yield driver
    except WebDriverException as e:
        # If Chrome is not installed, skip all tests
        if "cannot find Chrome binary" in str(e):
            pytest.skip("Chrome browser not installed. Please install Chrome to run Selenium tests.")
        else:
            # Re-raise other WebDriver exceptions
            raise
    except Exception as e:
        # Handle any other exceptions
        pytest.skip(f"Failed to initialize WebDriver: {str(e)}")
    finally:
        if driver:
            driver.quit()

@pytest.fixture(scope="session")
def base_url():
    return os.getenv('BASE_URL', 'http://localhost:5173')

@pytest.fixture(scope="session")
def api_base_url():
    return os.getenv('API_BASE_URL', 'http://localhost:8080')

@pytest.fixture
def wait(driver):
    return WebDriverWait(driver, 10)

@pytest.fixture(scope="session")
def admin_credentials():
    return {
        'email': os.getenv('ADMIN_EMAIL', 'admin@example.com'),
        'password': os.getenv('ADMIN_PASSWORD', 'admin123')
    }

@pytest.fixture(scope="session")
def user_credentials():
    return {
        'email': os.getenv('USER_EMAIL', 'user@example.com'),
        'password': os.getenv('USER_PASSWORD', 'user123')
    }
