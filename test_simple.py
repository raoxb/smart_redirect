#!/usr/bin/env python3
"""
简单的重定向测试脚本 - 展示核心功能
"""
import requests
import time
import json

BASE_URL = "http://103.14.79.22:8080"
ADMIN_URL = "http://103.14.79.22:3003"

def test_redirect(link_id="df7fca", ip="192.168.1.1"):
    """测试单次重定向"""
    url = f"{BASE_URL}/v1/bu01/{link_id}"
    headers = {"X-Real-IP": ip}
    
    print(f"\n📍 测试重定向: {url}")
    print(f"   客户端IP: {ip}")
    
    try:
        resp = requests.get(url, headers=headers, allow_redirects=False)
        if resp.status_code == 302:
            target = resp.headers.get('Location', 'N/A')
            print(f"   ✅ 重定向成功!")
            print(f"   目标URL: {target}")
            return True
        else:
            print(f"   ❌ 错误: {resp.status_code} - {resp.text}")
            return False
    except Exception as e:
        print(f"   ❌ 请求失败: {e}")
        return False

def test_ip_memory():
    """测试IP记忆功能"""
    print("\n🧠 测试IP记忆功能")
    print("   同一IP多次访问应分配到不同目标...")
    
    ip = "10.10.10.10"
    targets = []
    
    for i in range(3):
        url = f"{BASE_URL}/v1/bu01/df7fca"
        headers = {"X-Real-IP": ip}
        resp = requests.get(url, headers=headers, allow_redirects=False)
        
        if resp.status_code == 302:
            target = resp.headers.get('Location', '')
            targets.append(target.split('?')[0])  # 只取URL部分
            print(f"   访问 #{i+1}: {target.split('?')[0]}")
        
        time.sleep(0.5)
    
    unique_targets = set(targets)
    print(f"   结果: {len(targets)}次访问，{len(unique_targets)}个不同目标")
    
def test_geographic():
    """测试地理位置定向"""
    print("\n🌍 测试地理位置定向")
    
    test_cases = [
        ("8.8.8.8", "US", "美国"),
        ("46.165.230.5", "DE", "德国"),
        ("114.114.114.114", "CN", "中国"),
    ]
    
    for ip, country, name in test_cases:
        url = f"{BASE_URL}/v1/bu01/df7fca"
        headers = {"X-Real-IP": ip}
        resp = requests.get(url, headers=headers, allow_redirects=False)
        
        if resp.status_code == 302:
            target = resp.headers.get('Location', '')
            print(f"   {name} IP ({ip}): → {target.split('?')[0]}")

def test_rate_limit():
    """测试速率限制"""
    print("\n🚦 测试速率限制")
    print("   快速发送多个请求...")
    
    ip = "192.192.192.192"
    success = 0
    blocked = 0
    
    for i in range(12):
        url = f"{BASE_URL}/v1/bu01/df7fca"
        headers = {"X-Real-IP": ip}
        resp = requests.get(url, headers=headers, allow_redirects=False)
        
        if resp.status_code == 302:
            success += 1
        elif resp.status_code == 429:
            blocked += 1
            if blocked == 1:
                print(f"   ⚠️  第{i+1}次请求被限流!")
    
    print(f"   结果: 成功{success}次，被限流{blocked}次")

def test_parameter_transform():
    """测试参数转换"""
    print("\n🔄 测试参数转换")
    
    url = f"{BASE_URL}/v1/bu01/df7fca?src=google&cmp=summer2024"
    headers = {"X-Real-IP": "1.2.3.4"}
    
    print(f"   原始参数: src=google&cmp=summer2024")
    
    resp = requests.get(url, headers=headers, allow_redirects=False)
    if resp.status_code == 302:
        target = resp.headers.get('Location', '')
        print(f"   转换后: {target.split('?')[1] if '?' in target else 'No params'}")

def show_stats():
    """显示统计信息"""
    print("\n📊 访问统计")
    print(f"   管理后台: {ADMIN_URL}")
    print("   用户名: admin")
    print("   密码: admin123")
    print("\n   您可以在管理后台查看:")
    print("   - 实时访问统计")
    print("   - 地理分布图表")
    print("   - 目标命中率")
    print("   - IP访问记录")

def main():
    print("=" * 60)
    print("🚀 Smart Redirect 系统功能测试")
    print("=" * 60)
    
    # 基础重定向测试
    test_redirect()
    
    # IP记忆测试
    test_ip_memory()
    
    # 地理位置测试
    test_geographic()
    
    # 参数转换测试
    test_parameter_transform()
    
    # 速率限制测试
    test_rate_limit()
    
    # 显示统计
    show_stats()
    
    print("\n" + "=" * 60)
    print("✅ 测试完成!")
    print("=" * 60)

if __name__ == "__main__":
    main()