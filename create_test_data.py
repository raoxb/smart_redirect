#!/usr/bin/env python3
"""
Create test data for stability testing
"""

import requests
import json
import time

# Configuration
BASE_URL = "http://localhost:8080"
JWT_TOKEN = None

def get_jwt_token():
    """Get JWT token for authentication"""
    global JWT_TOKEN
    
    # Create admin user first
    admin_data = {
        "username": "admin",
        "password": "admin123",
        "email": "admin@example.com"
    }
    
    try:
        # Try to register admin user
        response = requests.post(f"{BASE_URL}/api/v1/auth/register", json=admin_data)
        if response.status_code not in [200, 201, 409]:  # 409 for user already exists
            print(f"Failed to register admin: {response.status_code} - {response.text}")
    except Exception as e:
        print(f"Error registering admin: {e}")
    
    # Login to get token
    login_data = {
        "username": "admin",
        "password": "admin123"
    }
    
    try:
        response = requests.post(f"{BASE_URL}/api/v1/auth/login", json=login_data)
        if response.status_code == 200:
            JWT_TOKEN = response.json().get('token')
            print(f"✅ Got JWT token: {JWT_TOKEN[:20]}...")
            return True
        else:
            print(f"❌ Failed to login: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"❌ Error logging in: {e}")
        return False

def create_test_links():
    """Create test links for stability testing"""
    if not JWT_TOKEN:
        print("❌ No JWT token available")
        return
    
    headers = {
        "Authorization": f"Bearer {JWT_TOKEN}",
        "Content-Type": "application/json"
    }
    
    # Test links configuration
    test_links = [
        {
            "business_unit": "marketing",
            "targets": [
                {"url": "https://example.com/page1", "weight": 50},
                {"url": "https://example.com/page2", "weight": 30},
                {"url": "https://example.com/page3", "weight": 20}
            ]
        },
        {
            "business_unit": "sales", 
            "targets": [
                {"url": "https://example.com/sales1", "weight": 40},
                {"url": "https://example.com/sales2", "weight": 60}
            ]
        },
        {
            "business_unit": "apps",
            "targets": [
                {"url": "https://example.com/app1", "weight": 25},
                {"url": "https://example.com/app2", "weight": 25},
                {"url": "https://example.com/app3", "weight": 25},
                {"url": "https://example.com/app4", "weight": 25}
            ]
        },
        {
            "business_unit": "blog",
            "targets": [
                {"url": "https://example.com/blog1", "weight": 80},
                {"url": "https://example.com/blog2", "weight": 20}
            ]
        }
    ]
    
    created_links = []
    
    for i, link_config in enumerate(test_links):
        try:
            # Create link
            response = requests.post(f"{BASE_URL}/api/v1/links", json=link_config, headers=headers)
            if response.status_code in [200, 201]:
                link_data = response.json()
                link_id = link_data.get('id')
                created_links.append(link_id)
                print(f"✅ Created link {i+1}: {link_id}")
                
                # Add some delay between requests
                time.sleep(0.5)
            else:
                print(f"❌ Failed to create link {i+1}: {response.status_code} - {response.text}")
        except Exception as e:
            print(f"❌ Error creating link {i+1}: {e}")
    
    return created_links

def main():
    print("🔧 Creating test data for stability testing...")
    
    # Get authentication token
    if not get_jwt_token():
        print("❌ Cannot proceed without authentication")
        return
    
    # Create test links
    created_links = create_test_links()
    
    if created_links:
        print(f"\n✅ Created {len(created_links)} test links:")
        for link_id in created_links:
            print(f"  - {link_id}")
        
        # Save link IDs to file for stress test
        with open('test_links.json', 'w') as f:
            json.dump(created_links, f, indent=2)
        
        print(f"\n📝 Test links saved to test_links.json")
        print(f"🚀 Ready to start stability testing!")
    else:
        print("❌ No test links were created")

if __name__ == "__main__":
    main()