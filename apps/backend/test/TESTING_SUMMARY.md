# Consent Manager Selenium Testing Setup

## Overview

This document provides a summary of the Selenium testing setup for the Consent Manager application. The testing framework has been successfully configured to support automated testing of the web application.

## Test Environment Setup

### Prerequisites

1. **Python 3.8 or higher** - Installed and verified
2. **Required Python packages** - Successfully installed via `requirements.txt`
3. **Web browser** - Chrome recommended (currently not installed)

### Installed Components

- Selenium WebDriver 4.27.1
- pytest 8.3.4
- pytest-html 4.1.1
- python-dotenv 1.0.1
- webdriver-manager 4.0.2
- requests 2.32.3
- pytest-xdist 3.6.1

### Directory Structure

```
test/
├── __init__.py
├── conftest.py
├── requirements.txt
├── .env.example
├── README.md
├── run_tests.py
├── setup_test_data.py
├── test_admin_features.py
├── test_user_features.py
├── test_api_integration.py
├── test_data_management.py
├── test_selenium_setup.py
├── test_report_template.md
└── TESTING_SUMMARY.md
```

## Test Suites

### 1. Admin Features (`test_admin_features.py`)

Tests for Data Fiduciary (admin) functionality:
- Admin login
- Creating consent forms
- Managing API keys
- Viewing integration scripts

### 2. User Features (`test_user_features.py`)

Tests for Data Principal (user) functionality:
- User login
- Viewing user consents
- Submitting consents
- Verifying access restrictions

### 3. API Integration (`test_api_integration.py`)

Tests for API key integration and third-party access:
- API key authentication
- Public consent form access via API
- Integration script generation
- User consent submission via API
- CORS header validation

### 4. Data Management (`test_data_management.py`)

Tests for test data setup and management:
- Creating test users
- Creating test consent forms
- Creating test API keys

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and update with appropriate values:

```bash
cp .env.example .env
```

Required variables:
- `BASE_URL` - Frontend application URL
- `API_BASE_URL` - API server URL
- `ADMIN_EMAIL` - Admin user email
- `ADMIN_PASSWORD` - Admin user password
- `USER_EMAIL` - Regular user email
- `USER_PASSWORD` - Regular user password

## Running Tests

### Using the Test Runner Script

```bash
python run_tests.py
```

### Using pytest Directly

```bash
# Run all tests
pytest

# Run specific test file
pytest test_admin_features.py

# Run tests with verbose output
pytest -v

# Run tests in parallel
pytest -n auto
```

### Test Data Setup

```bash
python setup_test_data.py
```

## Current Status

✅ Selenium testing framework configured
✅ All required Python packages installed
❌ Chrome browser not installed (tests will skip)
✅ Test directory structure created
✅ All test suites implemented
✅ Configuration files created
✅ Documentation completed

## Next Steps

1. Install Chrome browser to enable Selenium tests
2. Update `.env` with actual credentials
3. Run `setup_test_data.py` to create test data
4. Execute test suites to validate functionality
5. Generate test reports

## Troubleshooting

### Common Issues

1. **Chrome not installed**: Tests will skip with a clear message
   - Solution: Install Chrome browser

2. **Server not running**: API tests will fail
   - Solution: Start the Consent Manager server

3. **Invalid credentials**: Authentication tests will fail
   - Solution: Update credentials in `.env` file

### Test Execution Issues

1. **Timeout errors**: Increase wait times in tests
2. **Element not found**: Update selectors based on UI changes
3. **Authentication failures**: Verify credentials and server status

## Reporting

Test results are displayed in the terminal and HTML reports can be generated:

```bash
pytest --html=report.html --self-contained-html
```

## Maintenance

- Update selectors when UI changes
- Add new test cases for new features
- Review and update test data periodically
- Monitor test execution times and optimize as needed
