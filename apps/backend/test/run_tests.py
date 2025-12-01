#!/usr/bin/env python3

import subprocess
import sys
import os

def run_tests():
    """Run all Selenium tests and generate reports"""
    print("Starting Selenium tests for Consent Manager...")
    
    # Ensure we're in the test directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    # Install dependencies if not already installed
    print("Installing dependencies...")
    try:
        subprocess.run([sys.executable, "-m", "pip", "install", "-r", "requirements.txt"], check=True)
    except subprocess.CalledProcessError:
        print("Warning: Could not install dependencies. Continuing with tests...")
    
    # Run tests with coverage and generate HTML report
    print("Running tests...")
    try:
        # Run tests with detailed output
        result = subprocess.run([
            sys.executable, "-m", "pytest", 
            "-v", 
            "--tb=short",
            "--html=report.html",
            "--self-contained-html",
            "--junitxml=report.xml"
        ])
        
        if result.returncode == 0:
            print("\nAll tests passed successfully!")
        else:
            print(f"\nSome tests failed with exit code {result.returncode}")
            
        # Display summary
        print("\nTest reports generated:")
        print("  - HTML report: report.html")
        print("  - XML report: report.xml")
        
        return result.returncode
        
    except subprocess.CalledProcessError as e:
        print(f"Error running tests: {e}")
        return e.returncode
    except FileNotFoundError:
        print("Error: pytest not found. Please install it with 'pip install pytest pytest-html'")
        return 1

def run_specific_tests(test_file):
    """Run a specific test file"""
    print(f"Running tests from {test_file}...")
    
    try:
        result = subprocess.run([
            sys.executable, "-m", "pytest", 
            test_file,
            "-v"
        ])
        
        return result.returncode
    except Exception as e:
        print(f"Error running tests: {e}")
        return 1

if __name__ == "__main__":
    if len(sys.argv) > 1:
        # Run specific test file
        exit_code = run_specific_tests(sys.argv[1])
    else:
        # Run all tests
        exit_code = run_tests()
    
    sys.exit(exit_code)
