#!/usr/bin/env python3
"""
System Health Monitor
系统健康状况监控脚本
"""
import requests
import time
import json
import psutil
import subprocess
from datetime import datetime

# 配置
BASE_URL = "http://103.14.79.22:8080"
TOKEN_FILE = "/tmp/jwt_token"
LOG_FILE = "system_monitor.log"

def get_token():
    """获取JWT令牌"""
    try:
        with open(TOKEN_FILE, 'r') as f:
            return f.read().strip()
    except:
        return None

def check_api_health():
    """检查API健康状况"""
    try:
        # 检查基本健康端点
        response = requests.get(f"{BASE_URL}/health", timeout=5)
        if response.status_code == 200:
            return {"status": "healthy", "response_time": response.elapsed.total_seconds() * 1000}
        else:
            return {"status": "unhealthy", "status_code": response.status_code}
    except Exception as e:
        return {"status": "error", "error": str(e)}

def check_database_health():
    """检查数据库健康状况"""
    try:
        # 使用docker命令检查PostgreSQL
        result = subprocess.run([
            "docker-compose", "exec", "-T", "postgres", 
            "psql", "-U", "postgres", "-d", "smart_redirect", 
            "-c", "SELECT 1;"
        ], capture_output=True, text=True, timeout=10)
        
        if result.returncode == 0:
            return {"status": "healthy"}
        else:
            return {"status": "unhealthy", "error": result.stderr}
    except Exception as e:
        return {"status": "error", "error": str(e)}

def check_redis_health():
    """检查Redis健康状况"""
    try:
        result = subprocess.run([
            "docker-compose", "exec", "-T", "redis", 
            "redis-cli", "ping"
        ], capture_output=True, text=True, timeout=10)
        
        if result.returncode == 0 and "PONG" in result.stdout:
            return {"status": "healthy"}
        else:
            return {"status": "unhealthy", "error": result.stderr}
    except Exception as e:
        return {"status": "error", "error": str(e)}

def get_system_metrics():
    """获取系统资源使用情况"""
    try:
        cpu_percent = psutil.cpu_percent(interval=1)
        memory = psutil.virtual_memory()
        disk = psutil.disk_usage('/')
        
        return {
            "cpu_percent": cpu_percent,
            "memory_percent": memory.percent,
            "memory_used_gb": memory.used / (1024**3),
            "memory_total_gb": memory.total / (1024**3),
            "disk_percent": disk.percent,
            "disk_used_gb": disk.used / (1024**3),
            "disk_total_gb": disk.total / (1024**3)
        }
    except Exception as e:
        return {"error": str(e)}

def get_api_stats():
    """获取API统计信息"""
    token = get_token()
    if not token:
        return {"error": "No auth token"}
    
    headers = {"Authorization": f"Bearer {token}"}
    
    try:
        # 获取系统统计
        response = requests.get(f"{BASE_URL}/api/v1/stats/system", 
                              headers=headers, timeout=10)
        if response.status_code == 200:
            return response.json()
        else:
            return {"error": f"API returned {response.status_code}"}
    except Exception as e:
        return {"error": str(e)}

def get_database_stats():
    """获取数据库统计信息"""
    try:
        # 获取表大小和行数
        query = """
        SELECT 
            schemaname,
            tablename,
            attname,
            n_distinct,
            most_common_vals
        FROM pg_stats 
        WHERE schemaname = 'public' 
        AND tablename IN ('links', 'targets', 'access_logs', 'users');
        """
        
        result = subprocess.run([
            "docker-compose", "exec", "-T", "postgres",
            "psql", "-U", "postgres", "-d", "smart_redirect",
            "-c", query
        ], capture_output=True, text=True, timeout=10)
        
        if result.returncode == 0:
            return {"status": "success", "output": result.stdout}
        else:
            return {"error": result.stderr}
    except Exception as e:
        return {"error": str(e)}

def log_message(message, level="INFO"):
    """记录日志消息"""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    log_entry = f"[{timestamp}] {level}: {message}\n"
    
    print(log_entry.strip())
    
    with open(LOG_FILE, "a") as f:
        f.write(log_entry)

def generate_health_report():
    """生成健康状况报告"""
    log_message("=" * 60)
    log_message("系统健康状况检查开始")
    
    # API健康检查
    api_health = check_api_health()
    log_message(f"API健康状况: {json.dumps(api_health, indent=2)}")
    
    # 数据库健康检查
    db_health = check_database_health()
    log_message(f"数据库健康状况: {json.dumps(db_health, indent=2)}")
    
    # Redis健康检查
    redis_health = check_redis_health()
    log_message(f"Redis健康状况: {json.dumps(redis_health, indent=2)}")
    
    # 系统资源
    system_metrics = get_system_metrics()
    log_message(f"系统资源使用: {json.dumps(system_metrics, indent=2)}")
    
    # API统计
    api_stats = get_api_stats()
    log_message(f"API统计信息: {json.dumps(api_stats, indent=2)}")
    
    # 检查警告条件
    warnings = []
    
    if api_health.get("status") != "healthy":
        warnings.append("API不健康")
    
    if db_health.get("status") != "healthy":
        warnings.append("数据库不健康")
    
    if redis_health.get("status") != "healthy":
        warnings.append("Redis不健康")
    
    if "cpu_percent" in system_metrics and system_metrics["cpu_percent"] > 80:
        warnings.append(f"CPU使用率过高: {system_metrics['cpu_percent']:.1f}%")
    
    if "memory_percent" in system_metrics and system_metrics["memory_percent"] > 85:
        warnings.append(f"内存使用率过高: {system_metrics['memory_percent']:.1f}%")
    
    if "disk_percent" in system_metrics and system_metrics["disk_percent"] > 90:
        warnings.append(f"磁盘使用率过高: {system_metrics['disk_percent']:.1f}%")
    
    if warnings:
        log_message("⚠️  检测到警告:", "WARNING")
        for warning in warnings:
            log_message(f"  - {warning}", "WARNING")
    else:
        log_message("✅ 所有检查通过，系统运行正常")
    
    log_message("系统健康状况检查完成")
    log_message("=" * 60)

def main():
    """主函数"""
    print("🔍 Smart Redirect 系统监控器")
    print("每5分钟检查一次系统健康状况")
    print("按Ctrl+C停止监控")
    print("-" * 60)
    
    try:
        while True:
            generate_health_report()
            print(f"⏰ 下次检查时间: {datetime.now().strftime('%H:%M:%S')} (5分钟后)")
            time.sleep(300)  # 5分钟检查一次
    except KeyboardInterrupt:
        print("\n🛑 监控已停止")

if __name__ == "__main__":
    main()