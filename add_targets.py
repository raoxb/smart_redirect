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

# 链接ID到目标的映射 (基于上面创建的实际link_id)
links_targets = {
    "e79e4f": [  # marketing/social
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
            "cap": 0,
            "countries": [],
            "static_params": {
                "utm_source": "redirect",
                "utm_medium": "social"
            }
        }
    ],
    "878471": [  # apps/mobile
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
    ],
    "8223fb": [  # sales/email
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
    ],
    "fffd75": [  # blog/social
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
}

def add_targets_to_link(link_id, targets):
    """为指定链接添加目标"""
    print(f"Adding targets to link: {link_id}")
    
    success_count = 0
    for i, target in enumerate(targets):
        print(f"  Adding target {i+1}: {target['url']}")
        
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
            f"{BASE_URL}/links/{link_id}/targets", 
            headers=headers, 
            json=target_payload
        )
        
        if response.status_code == 201:
            print(f"  ✓ Target added successfully")
            success_count += 1
        else:
            print(f"  ✗ Failed to add target: {response.text}")
        
        time.sleep(0.5)  # Small delay
    
    print(f"Added {success_count}/{len(targets)} targets to {link_id}\n")
    return success_count

def main():
    print("Adding targets to created links...")
    print("=" * 50)
    
    total_targets = 0
    total_success = 0
    
    for link_id, targets in links_targets.items():
        total_targets += len(targets)
        total_success += add_targets_to_link(link_id, targets)
        time.sleep(1)  # Delay between links
    
    print(f"Successfully added {total_success}/{total_targets} targets across {len(links_targets)} links")
    print("Target creation completed!")

if __name__ == "__main__":
    main()