#!/usr/bin/env python3
import requests
import time
import random
from collections import defaultdict

# 测试配置
BASE_URL = "http://103.14.79.22:8080"
LINK_ID = "df7fca"
BU = "bu01"

# 不同地区的模拟IP
TEST_IPS = {
    "US": ["8.8.8.8", "1.1.1.1", "208.67.222.222"],
    "UK": ["81.2.69.142", "81.2.69.143", "81.2.69.144"],
    "DE": ["46.165.230.5", "46.165.230.6", "46.165.230.7"],
    "CN": ["114.114.114.114", "223.5.5.5", "119.29.29.29"],
    "JP": ["203.119.101.61", "203.119.101.62", "203.119.101.63"],
}

def test_redirect(ip, params=None):
    """测试单个重定向请求"""
    url = f"{BASE_URL}/v1/{BU}/{LINK_ID}"
    headers = {
        "X-Real-IP": ip,
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }
    
    try:
        # 发送请求，不跟随重定向
        response = requests.get(url, headers=headers, params=params, allow_redirects=False)
        
        if response.status_code == 302:
            target_url = response.headers.get('Location', 'N/A')
            print(f"✓ IP {ip}: Redirected to {target_url}")
            return {"success": True, "target": target_url, "status": response.status_code}
        else:
            print(f"✗ IP {ip}: Status {response.status_code} - {response.text}")
            return {"success": False, "status": response.status_code, "error": response.text}
    except Exception as e:
        print(f"✗ IP {ip}: Error - {str(e)}")
        return {"success": False, "error": str(e)}

def test_ip_memory():
    """测试IP记忆功能 - 同一IP多次访问应该得到不同的目标"""
    print("\n=== Testing IP Memory Feature ===")
    test_ip = "192.168.1.100"
    targets = []
    
    for i in range(5):
        print(f"\nVisit #{i+1} from IP {test_ip}:")
        result = test_redirect(test_ip, {"src": "test", "cmp": f"visit{i+1}"})
        if result["success"]:
            targets.append(result["target"])
        time.sleep(1)
    
    unique_targets = set(targets)
    print(f"\nIP Memory Test Result:")
    print(f"Total visits: {len(targets)}")
    print(f"Unique targets: {len(unique_targets)}")
    print(f"Targets: {list(unique_targets)}")
    
    return len(unique_targets) > 1

def test_geographic_distribution():
    """测试地理位置分发"""
    print("\n=== Testing Geographic Distribution ===")
    results = defaultdict(list)
    
    for country, ips in TEST_IPS.items():
        print(f"\nTesting {country} IPs:")
        for ip in ips:
            result = test_redirect(ip, {"network": "test", "geo": country})
            if result["success"]:
                results[country].append(result["target"])
            time.sleep(0.5)
    
    print("\n=== Geographic Distribution Results ===")
    for country, targets in results.items():
        unique_targets = set(targets)
        print(f"{country}: {len(targets)} requests → {len(unique_targets)} unique targets")
        for target in unique_targets:
            count = targets.count(target)
            print(f"  - {target}: {count} times")

def test_rate_limiting():
    """测试速率限制"""
    print("\n=== Testing Rate Limiting ===")
    test_ip = "10.0.0.1"
    
    # 快速发送多个请求
    for i in range(15):
        print(f"Request #{i+1}:", end=" ")
        result = test_redirect(test_ip, {"test": "ratelimit"})
        if not result["success"] and "blocked" in result.get("error", "").lower():
            print(f"\nRate limit triggered after {i} requests")
            return True
        time.sleep(0.1)
    
    print("\nNo rate limit triggered")
    return False

def test_parameter_transformation():
    """测试参数转换功能"""
    print("\n=== Testing Parameter Transformation ===")
    test_params = {
        "src": "google",
        "cmp": "summer2024",
        "utm_source": "facebook",
        "click_id": "abc123"
    }
    
    result = test_redirect("1.1.1.1", test_params)
    if result["success"]:
        print(f"Original params: {test_params}")
        print(f"Redirected to: {result['target']}")
        # 检查URL中的参数
        if "?" in result["target"]:
            query_string = result["target"].split("?")[1]
            print(f"Query string: {query_string}")

def test_load():
    """负载测试 - 模拟多个用户并发访问"""
    print("\n=== Load Testing (50 requests) ===")
    start_time = time.time()
    success_count = 0
    target_distribution = defaultdict(int)
    
    for i in range(50):
        # 随机选择一个IP
        country = random.choice(list(TEST_IPS.keys()))
        ip = random.choice(TEST_IPS[country])
        
        result = test_redirect(ip, {"load": "test", "req": i})
        if result["success"]:
            success_count += 1
            target_distribution[result["target"]] += 1
        
        # 模拟真实用户间隔
        time.sleep(random.uniform(0.1, 0.3))
    
    end_time = time.time()
    duration = end_time - start_time
    
    print(f"\n=== Load Test Results ===")
    print(f"Total requests: 50")
    print(f"Successful: {success_count}")
    print(f"Failed: {50 - success_count}")
    print(f"Success rate: {success_count/50*100:.1f}%")
    print(f"Total time: {duration:.2f}s")
    print(f"Avg response time: {duration/50*1000:.1f}ms")
    print(f"\nTarget distribution:")
    for target, count in sorted(target_distribution.items(), key=lambda x: x[1], reverse=True):
        print(f"  - {target}: {count} requests ({count/success_count*100:.1f}%)")

def main():
    print("=" * 60)
    print("Smart Redirect System Test")
    print("=" * 60)
    print(f"Testing URL: {BASE_URL}/v1/{BU}/{LINK_ID}")
    print("=" * 60)
    
    # 运行各项测试
    test_ip_memory()
    test_geographic_distribution()
    test_parameter_transformation()
    test_rate_limiting()
    test_load()
    
    print("\n" + "=" * 60)
    print("Test completed!")
    print("Check the admin panel for statistics:")
    print("http://103.14.79.22:3003/")
    print("=" * 60)

if __name__ == "__main__":
    main()