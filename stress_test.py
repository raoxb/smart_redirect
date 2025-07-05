#!/usr/bin/env python3
"""
Smart Redirect Stress Test Script
持续1-2天访问多个短链接进行稳定性和性能测试
"""
import requests
import random
import time
import threading
import json
import sys
from datetime import datetime, timedelta
import signal
import os

# 测试配置
BASE_URL = "http://103.14.79.22:8080/v1"
BUSINESS_UNITS = ["bu01", "marketing", "apps", "sales", "blog"]
LINK_IDS = ["df7fca", "e79e4f", "878471", "8223fb", "fffd75"]

# 模拟不同地区的IP地址池
IP_POOLS = {
    "US": ["8.8.8.8", "1.1.1.1", "208.67.222.222", "208.67.220.220"],
    "CN": ["223.5.5.5", "223.6.6.6", "114.114.114.114", "119.29.29.29"],
    "GB": ["81.2.69.142", "81.2.69.143", "81.2.69.144", "1.1.1.2"],
    "DE": ["46.165.230.7", "46.165.230.8", "194.25.0.68", "194.25.0.69"],
    "AU": ["1.1.1.3", "139.130.4.5", "203.112.2.4", "203.112.2.5"],
    "CA": ["99.250.191.7", "99.250.191.8", "142.150.191.7", "142.150.191.8"],
    "FR": ["212.27.40.240", "212.27.40.241", "80.67.169.12", "80.67.169.13"],
    "IT": ["151.11.48.50", "151.11.48.51", "193.70.152.25", "193.70.152.26"],
    "LOCAL": ["127.0.0.1", "192.168.1.100", "10.0.0.1", "172.16.0.1"]
}

# User Agent池
USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
    "Mozilla/5.0 (iPad; CPU OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
    "Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Mobile Safari/537.36"
]

# Referer池
REFERERS = [
    "https://www.google.com/search?q=example",
    "https://www.facebook.com/",
    "https://twitter.com/",
    "https://www.linkedin.com/",
    "https://www.reddit.com/",
    "https://www.youtube.com/",
    "https://www.instagram.com/",
    "https://t.co/abc123",
    "https://bit.ly/xyz789",
    ""  # Direct access
]

# 全局统计
stats = {
    "total_requests": 0,
    "successful_requests": 0,
    "failed_requests": 0,
    "avg_response_time": 0,
    "errors": {},
    "start_time": None,
    "last_report": None
}

# 停止标志
stop_flag = threading.Event()

def signal_handler(signum, frame):
    """处理中断信号"""
    print(f"\n🛑 收到信号 {signum}，准备停止测试...")
    stop_flag.set()

def get_random_ip():
    """随机选择一个IP地址"""
    country = random.choice(list(IP_POOLS.keys()))
    ip = random.choice(IP_POOLS[country])
    return ip, country

def make_request(bu, link_id, ip, country):
    """发送单个请求"""
    url = f"{BASE_URL}/{bu}/{link_id}"
    
    # 随机添加参数
    params = {}
    if random.random() < 0.3:  # 30%概率添加参数
        params = {
            "source": random.choice(["social", "email", "direct", "search"]),
            "campaign": random.choice(["summer", "winter", "spring", "fall"]),
            "ref": random.choice(["facebook", "twitter", "google", "linkedin"])
        }
    
    headers = {
        "X-Real-IP": ip,
        "User-Agent": random.choice(USER_AGENTS),
        "Referer": random.choice(REFERERS),
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
        "Accept-Encoding": "gzip, deflate",
        "Connection": "keep-alive",
        "Upgrade-Insecure-Requests": "1"
    }
    
    start_time = time.time()
    try:
        response = requests.get(url, headers=headers, params=params, 
                              allow_redirects=False, timeout=10)
        end_time = time.time()
        response_time = (end_time - start_time) * 1000  # ms
        
        with threading.Lock():
            stats["total_requests"] += 1
            if 300 <= response.status_code < 400:  # 重定向成功
                stats["successful_requests"] += 1
            else:
                stats["failed_requests"] += 1
                error_key = f"{response.status_code}"
                stats["errors"][error_key] = stats["errors"].get(error_key, 0) + 1
            
            # 更新平均响应时间
            current_avg = stats["avg_response_time"]
            total = stats["total_requests"]
            stats["avg_response_time"] = ((current_avg * (total - 1)) + response_time) / total
        
        return response.status_code, response_time, len(response.content)
        
    except requests.exceptions.RequestException as e:
        end_time = time.time()
        response_time = (end_time - start_time) * 1000
        
        with threading.Lock():
            stats["total_requests"] += 1
            stats["failed_requests"] += 1
            error_key = f"Exception: {type(e).__name__}"
            stats["errors"][error_key] = stats["errors"].get(error_key, 0) + 1
        
        return 0, response_time, 0

def worker_thread(thread_id, duration_hours):
    """工作线程"""
    print(f"🚀 线程 {thread_id} 开始运行")
    
    end_time = datetime.now() + timedelta(hours=duration_hours)
    request_count = 0
    
    while not stop_flag.is_set() and datetime.now() < end_time:
        # 随机选择链接
        bu = random.choice(BUSINESS_UNITS)
        link_id = random.choice(LINK_IDS)
        ip, country = get_random_ip()
        
        # 发送请求
        status_code, response_time, content_length = make_request(bu, link_id, ip, country)
        request_count += 1
        
        # 随机延迟 (0.1 - 5 秒，模拟真实用户)
        delay = random.uniform(0.1, 5.0)
        time.sleep(delay)
        
        # 每100个请求输出一次状态
        if request_count % 100 == 0:
            print(f"🔄 线程 {thread_id}: 已完成 {request_count} 个请求")
    
    print(f"✅ 线程 {thread_id} 完成，共处理 {request_count} 个请求")

def print_stats():
    """打印统计信息"""
    now = datetime.now()
    if stats["start_time"]:
        elapsed = (now - stats["start_time"]).total_seconds()
        rps = stats["total_requests"] / elapsed if elapsed > 0 else 0
    else:
        elapsed = 0
        rps = 0
    
    success_rate = (stats["successful_requests"] / stats["total_requests"] * 100) if stats["total_requests"] > 0 else 0
    
    print(f"\n📊 === 测试统计 (运行时间: {elapsed/3600:.1f}小时) ===")
    print(f"总请求数: {stats['total_requests']:,}")
    print(f"成功请求: {stats['successful_requests']:,} ({success_rate:.1f}%)")
    print(f"失败请求: {stats['failed_requests']:,}")
    print(f"平均响应时间: {stats['avg_response_time']:.1f}ms")
    print(f"请求速率: {rps:.1f} req/s")
    
    if stats["errors"]:
        print("\n❌ 错误统计:")
        for error, count in stats["errors"].items():
            print(f"  {error}: {count}")
    
    print("=" * 50)

def monitor_thread():
    """监控线程，定期输出统计信息"""
    while not stop_flag.is_set():
        time.sleep(300)  # 每5分钟报告一次
        if not stop_flag.is_set():
            print_stats()

def save_stats_to_file():
    """保存统计信息到文件"""
    filename = f"stress_test_results_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    
    result = {
        "test_config": {
            "base_url": BASE_URL,
            "business_units": BUSINESS_UNITS,
            "link_ids": LINK_IDS,
            "ip_pools": IP_POOLS
        },
        "stats": stats,
        "test_duration_hours": (datetime.now() - stats["start_time"]).total_seconds() / 3600 if stats["start_time"] else 0
    }
    
    with open(filename, 'w') as f:
        json.dump(result, f, indent=2, default=str)
    
    print(f"📄 测试结果已保存到: {filename}")

def main():
    if len(sys.argv) < 2:
        print("使用方法: python3 stress_test.py <测试小时数> [线程数]")
        print("示例: python3 stress_test.py 24 5  # 运行24小时，5个线程")
        sys.exit(1)
    
    duration_hours = float(sys.argv[1])
    num_threads = int(sys.argv[2]) if len(sys.argv) > 2 else 3
    
    print("🧪 Smart Redirect 稳定性测试")
    print("=" * 50)
    print(f"测试时长: {duration_hours} 小时")
    print(f"并发线程: {num_threads}")
    print(f"目标URL: {BASE_URL}")
    print(f"测试链接: {LINK_IDS}")
    print("=" * 50)
    
    # 设置信号处理
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    # 初始化统计
    stats["start_time"] = datetime.now()
    stats["last_report"] = datetime.now()
    
    # 启动监控线程
    monitor = threading.Thread(target=monitor_thread, daemon=True)
    monitor.start()
    
    # 启动工作线程
    threads = []
    for i in range(num_threads):
        thread = threading.Thread(target=worker_thread, args=(i+1, duration_hours))
        thread.start()
        threads.append(thread)
        time.sleep(0.5)  # 错开启动时间
    
    print(f"🎯 测试开始！预计结束时间: {datetime.now() + timedelta(hours=duration_hours)}")
    
    try:
        # 等待所有工作线程完成
        for thread in threads:
            thread.join()
    except KeyboardInterrupt:
        print("\n⏹️  检测到中断，等待线程安全退出...")
        stop_flag.set()
        for thread in threads:
            thread.join(timeout=10)
    
    print("\n🏁 测试完成!")
    print_stats()
    save_stats_to_file()

if __name__ == "__main__":
    main()