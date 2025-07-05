#!/bin/bash

# Smart Redirect Long-term Stability Test
# 启动长期稳定性测试

echo "🚀 Smart Redirect 长期稳定性测试"
echo "=================================="

# 检查依赖
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 未安装"
    exit 1
fi

if ! pip3 list | grep -q requests; then
    echo "📦 安装 requests 库..."
    pip3 install requests psutil
fi

# 设置测试参数
DURATION_HOURS=${1:-48}  # 默认48小时
THREADS=${2:-3}          # 默认3个线程

echo "⏱️  测试时长: ${DURATION_HOURS} 小时"
echo "🧵 并发线程: ${THREADS}"
echo "📍 目标服务: http://103.14.79.22:8080"

# 创建日志目录
mkdir -p logs
LOG_DIR="logs/$(date '+%Y%m%d_%H%M%S')"
mkdir -p $LOG_DIR

echo "📂 日志目录: $LOG_DIR"

# 启动压力测试
echo "🎯 启动压力测试..."
nohup python3 stress_test.py $DURATION_HOURS $THREADS > $LOG_DIR/stress_test.log 2>&1 &
STRESS_PID=$!
echo "   压力测试PID: $STRESS_PID"

# 启动系统监控
echo "🔍 启动系统监控..."
nohup python3 monitor_system.py > $LOG_DIR/system_monitor.log 2>&1 &
MONITOR_PID=$!
echo "   监控PID: $MONITOR_PID"

# 保存PID到文件
echo $STRESS_PID > $LOG_DIR/stress_test.pid
echo $MONITOR_PID > $LOG_DIR/monitor.pid

# 创建停止脚本
cat > $LOG_DIR/stop_test.sh << EOF
#!/bin/bash
echo "🛑 停止长期测试..."

if [ -f stress_test.pid ]; then
    STRESS_PID=\$(cat stress_test.pid)
    echo "停止压力测试 (PID: \$STRESS_PID)..."
    kill \$STRESS_PID 2>/dev/null
    rm stress_test.pid
fi

if [ -f monitor.pid ]; then
    MONITOR_PID=\$(cat monitor.pid)
    echo "停止监控 (PID: \$MONITOR_PID)..."
    kill \$MONITOR_PID 2>/dev/null
    rm monitor.pid
fi

echo "✅ 测试已停止"
EOF

chmod +x $LOG_DIR/stop_test.sh

echo ""
echo "✅ 长期测试已启动！"
echo ""
echo "📋 管理命令:"
echo "   查看压力测试日志: tail -f $LOG_DIR/stress_test.log"
echo "   查看系统监控日志: tail -f $LOG_DIR/system_monitor.log"
echo "   停止测试: $LOG_DIR/stop_test.sh"
echo ""
echo "📊 实时监控:"
echo "   后台管理界面: http://103.14.79.22:3001"
echo "   访问日志页面: http://103.14.79.22:3001/access-logs"
echo ""
echo "⏰ 预计完成时间: $(date -d "+${DURATION_HOURS} hours" '+%Y-%m-%d %H:%M:%S')"
echo ""
echo "🔄 测试将持续运行，生成详细的性能和稳定性报告..."