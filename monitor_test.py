#!/usr/bin/env python3
"""
Monitor running stability test and display statistics
"""

import time
import json
import glob
import os
from datetime import datetime

def get_latest_results():
    """Get the latest test results file"""
    files = glob.glob('stress_test_results_*.json')
    if not files:
        return None
    return max(files, key=os.path.getmtime)

def display_stats():
    """Display current test statistics"""
    results_file = get_latest_results()
    if not results_file:
        print("No test results found yet...")
        return
    
    try:
        with open(results_file, 'r') as f:
            data = json.load(f)
        
        stats = data.get('stats', {})
        config = data.get('test_config', {})
        
        print("\n" + "="*60)
        print(f"STABILITY TEST MONITOR - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*60)
        
        print(f"\nTest Configuration:")
        print(f"  Base URL: {config.get('base_url', 'N/A')}")
        print(f"  Links tested: {len(config.get('link_ids', []))}")
        print(f"  IP pools: {len(config.get('ip_pools', {}))}")
        
        print(f"\nTest Statistics:")
        print(f"  Total requests: {stats.get('total_requests', 0):,}")
        print(f"  Successful: {stats.get('successful_requests', 0):,}")
        print(f"  Failed: {stats.get('failed_requests', 0):,}")
        
        success_rate = 0
        if stats.get('total_requests', 0) > 0:
            success_rate = (stats.get('successful_requests', 0) / stats.get('total_requests', 0)) * 100
        print(f"  Success rate: {success_rate:.2f}%")
        
        print(f"  Avg response time: {stats.get('avg_response_time', 0):.2f}ms")
        
        print(f"\nTest Duration: {data.get('test_duration_hours', 0):.2f} hours")
        
        if stats.get('errors'):
            print(f"\nErrors:")
            for error, count in stats['errors'].items():
                print(f"  {error}: {count}")
        
        print("\n" + "-"*60)
        
    except Exception as e:
        print(f"Error reading results: {e}")

def monitor_loop():
    """Main monitoring loop"""
    print("Starting test monitor (Press Ctrl+C to stop)...")
    
    try:
        while True:
            os.system('clear')  # Clear screen for better visibility
            display_stats()
            
            # Check process status
            pid_check = os.popen('ps aux | grep stress_test.py | grep -v grep').read()
            if pid_check:
                print("\nStress test process: RUNNING")
            else:
                print("\nStress test process: NOT RUNNING")
            
            time.sleep(30)  # Update every 30 seconds
            
    except KeyboardInterrupt:
        print("\nMonitor stopped.")

if __name__ == "__main__":
    monitor_loop()