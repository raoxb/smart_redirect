#!/usr/bin/env python3
"""
Stability test runner for Smart Redirect system
Runs continuous load testing for specified duration
"""

import subprocess
import time
import signal
import sys
import os
from datetime import datetime, timedelta

def signal_handler(sig, frame):
    print('\nInterrupted! Stopping test...')
    sys.exit(0)

def run_stability_test(duration_hours=48):
    """Run stability test for specified duration"""
    
    print(f"Starting {duration_hours}-hour stability test...")
    print(f"Test will run until: {datetime.now() + timedelta(hours=duration_hours)}")
    
    # Register signal handler
    signal.signal(signal.SIGINT, signal_handler)
    
    # Start stress test with proper parameters
    # stress_test.py expects: <duration_hours> [num_threads]
    cmd = [
        'python3', 'stress_test.py',
        str(duration_hours),  # duration in hours
        '4'  # number of threads
    ]
    
    # Create log directory
    log_dir = f"logs/stability_test_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
    os.makedirs(log_dir, exist_ok=True)
    
    # Open log file
    log_file = open(f"{log_dir}/stress_test.log", 'w')
    
    try:
        # Start the stress test process
        process = subprocess.Popen(
            cmd,
            stdout=log_file,
            stderr=subprocess.STDOUT,
            universal_newlines=True
        )
        
        print(f"Stress test started with PID: {process.pid}")
        print(f"Logs are being written to: {log_dir}/stress_test.log")
        
        # Monitor the process
        start_time = time.time()
        while process.poll() is None:
            elapsed_hours = (time.time() - start_time) / 3600
            remaining_hours = duration_hours - elapsed_hours
            
            print(f"\rProgress: {elapsed_hours:.1f}/{duration_hours} hours ({remaining_hours:.1f} hours remaining)", end='', flush=True)
            
            # Check every minute
            time.sleep(60)
            
            # Also print to log file
            log_file.write(f"\n[{datetime.now()}] Progress: {elapsed_hours:.1f}/{duration_hours} hours\n")
            log_file.flush()
        
        # Process completed
        print(f"\n\nStress test completed after {elapsed_hours:.1f} hours")
        
    except KeyboardInterrupt:
        print("\n\nTest interrupted by user")
        process.terminate()
        process.wait()
    finally:
        log_file.close()
        print(f"Test logs saved to: {log_dir}/stress_test.log")

if __name__ == "__main__":
    # Run 48-hour test by default
    duration = 48
    if len(sys.argv) > 1:
        duration = int(sys.argv[1])
    
    run_stability_test(duration)