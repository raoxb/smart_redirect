#!/usr/bin/env python3
import requests
import json
import time

# 配置
BASE_URL = "http://103.14.79.22:8080/api/v1"
TOKEN_FILE = "/tmp/jwt_token"

# 读取JWT令牌
with open(TOKEN_FILE, 'r') as f:
    token = f.read().strip()

headers = {
    "Authorization": f"Bearer {token}",
    "Content-Type": "application/json"
}

# 测试链接配置
test_links = [
    {
        "link_id": "test001",
        "business_unit": "marketing", 
        "network": "social",
        "total_cap": 50000,
        "backup_url": "https://www.example.com/backup",
        "targets": [
            {
                "url": "https://shop.example.com/product/shoes",
                "weight": 40,
                "cap": 20000,
                "countries": ["US", "CA", "UK"],
                "static_params": {
                    "utm_source": "redirect",
                    "utm_medium": "social",
                    "utm_campaign": "shoes"
                }
            },
            {
                "url": "https://store.example.com/footwear",
                "weight": 35,
                "cap": 15000,
                "countries": ["DE", "FR", "IT", "ES"],
                "static_params": {
                    "utm_source": "redirect",
                    "utm_medium": "social",
                    "utm_campaign": "europe_shoes"
                }
            },
            {
                "url": "https://market.example.com/shoes",
                "weight": 25,
                "cap": 0,  # No cap
                "countries": [],  # All other countries
                "static_params": {
                    "utm_source": "redirect",
                    "utm_medium": "social"
                }
            }
        ]
    },
    {
        "link_id": "mobile01",
        "business_unit": "apps",
        "network": "mobile",
        "total_cap": 30000,
        "backup_url": "https://www.example.com/mobile",
        "targets": [
            {
                "url": "https://apps.apple.com/app/example",
                "weight": 50,
                "cap": 15000,
                "countries": ["US", "CA", "UK", "AU"],
                "static_params": {
                    "platform": "ios",
                    "source": "redirect"
                }
            },
            {
                "url": "https://play.google.com/store/apps/details?id=com.example",
                "weight": 50,
                "cap": 15000,
                "countries": [],
                "static_params": {
                    "platform": "android",
                    "source": "redirect"
                }
            }
        ]
    },
    {
        "link_id": "promo02",
        "business_unit": "sales",
        "network": "email",
        "total_cap": 100000,
        "backup_url": "https://www.example.com/sale",
        "targets": [
            {
                "url": "https://sale.example.com/summer",
                "weight": 60,
                "cap": 60000,
                "countries": ["US", "CA"],
                "param_mapping": {
                    "ref": "referrer",
                    "src": "source"
                },
                "static_params": {
                    "campaign": "summer_sale",
                    "medium": "email"
                }
            },
            {
                "url": "https://discount.example.com/offers",
                "weight": 40,
                "cap": 40000,
                "countries": ["GB", "IE", "AU", "NZ"],
                "static_params": {
                    "campaign": "summer_sale_intl",
                    "medium": "email"
                }
            }
        ]
    },
    {
        "link_id": "content",
        "business_unit": "blog",
        "network": "social",
        "total_cap": 25000,
        "backup_url": "https://blog.example.com",
        "targets": [
            {
                "url": "https://blog.example.com/article/tech-trends",
                "weight": 70,
                "cap": 17500,
                "countries": ["US", "CA", "UK", "AU"],
                "static_params": {
                    "content_type": "article",
                    "topic": "tech"
                }
            },
            {
                "url": "https://news.example.com/technology",
                "weight": 30,
                "cap": 7500,
                "countries": [],
                "static_params": {
                    "content_type": "news",
                    "topic": "tech"
                }
            }
        ]
    },
    {
        "link_id": "event01",
        "business_unit": "events",
        "network": "paid",
        "total_cap": 15000,
        "backup_url": "https://events.example.com",
        "targets": [
            {
                "url": "https://eventbrite.com/e/example-conference",
                "weight": 80,
                "cap": 12000,
                "countries": ["US", "CA"],
                "static_params": {
                    "event": "conference_2025",
                    "source": "redirect"
                }
            },
            {
                "url": "https://meetup.com/example-group",
                "weight": 20,
                "cap": 3000,
                "countries": [],
                "static_params": {
                    "event": "meetup",
                    "source": "redirect"
                }
            }
        ]
    }
]

def create_link_with_targets(link_data):
    """创建链接及其目标"""
    print(f"Creating link: {link_data['link_id']}")
    
    # 准备链接数据
    link_payload = {
        "link_id": link_data["link_id"],
        "business_unit": link_data["business_unit"],
        "network": link_data["network"],
        "total_cap": link_data["total_cap"],
        "backup_url": link_data["backup_url"],
        "is_active": True
    }
    
    # 创建链接
    response = requests.post(f"{BASE_URL}/links", headers=headers, json=link_payload)
    if response.status_code != 201:
        print(f"Failed to create link {link_data['link_id']}: {response.text}")
        return False
    
    print(f"Link {link_data['link_id']} created successfully")
    
    # 创建目标
    for i, target in enumerate(link_data["targets"]):
        print(f"  Creating target {i+1}: {target['url']}")
        
        target_payload = {
            "url": target["url"],
            "weight": target["weight"],
            "cap": target["cap"],
            "countries": target["countries"],
            "param_mapping": target.get("param_mapping", {}),
            "static_params": target["static_params"],
            "is_active": True
        }
        
        response = requests.post(
            f"{BASE_URL}/links/{link_data['link_id']}/targets", 
            headers=headers, 
            json=target_payload
        )
        
        if response.status_code != 201:
            print(f"  Failed to create target: {response.text}")
        else:
            print(f"  Target created successfully")
        
        time.sleep(0.5)  # Small delay between requests
    
    print(f"Completed link {link_data['link_id']}\n")
    return True

def main():
    print("Creating test links and targets...")
    print("=" * 50)
    
    created_count = 0
    for link_data in test_links:
        if create_link_with_targets(link_data):
            created_count += 1
        time.sleep(1)  # Delay between links
    
    print(f"Successfully created {created_count}/{len(test_links)} links")
    print("Test links creation completed!")

if __name__ == "__main__":
    main()