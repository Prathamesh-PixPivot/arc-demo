#!/usr/bin/env python3

import subprocess
import sys
import os

def check_chrome_installed():
    """Check if Chrome is installed on the system"""
    try:
        # Try to find Chrome on Windows
        if sys.platform == "win32":
            # Check in common installation paths
            chrome_paths = [
                "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
                "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
                "C:\\Users\\{}\\AppData\\Local\\Google\\Chrome\\Application\\chrome.exe".format(os.getenv('USERNAME'))
            ]
            
            for path in chrome_paths:
                if os.path.exists(path):
                    return True, path
            
            # Try using where command
            result = subprocess.run(["where", "chrome.exe"], capture_output=True, text=True)
            if result.returncode == 0 and result.stdout:
                return True, result.stdout.strip().split('\n')[0]
                
        # Try to find Chrome on macOS
        elif sys.platform == "darwin":
            chrome_path = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
            if os.path.exists(chrome_path):
                return True, chrome_path
                
        # Try to find Chrome on Linux
        elif sys.platform.startswith("linux"):
            result = subprocess.run(["which", "google-chrome"], capture_output=True, text=True)
            if result.returncode == 0 and result.stdout:
                return True, result.stdout.strip()
                
            result = subprocess.run(["which", "chromium-browser"], capture_output=True, text=True)
            if result.returncode == 0 and result.stdout:
                return True, result.stdout.strip()
                
        # Try running chrome command directly
        result = subprocess.run(["chrome", "--version"], capture_output=True, text=True)
        if result.returncode == 0:
            return True, "chrome"
            
    except Exception as e:
        pass
        
    return False, None

def main():
    """Main function to check Chrome installation"""
    print("Checking for Chrome installation...")
    
    installed, path = check_chrome_installed()
    
    if installed:
        print("Chrome is installed at: {}".format(path))
        print("You can now run Selenium tests.")
        return 0
    else:
        print("Chrome is not installed on this system.")
        print("\nTo run Selenium tests, please install Chrome from:")
        print("https://www.google.com/chrome/")
        print("\nAfter installing Chrome, you can run the Selenium tests using:")
        print("python run_tests.py")
        return 1

if __name__ == "__main__":
    sys.exit(main())
