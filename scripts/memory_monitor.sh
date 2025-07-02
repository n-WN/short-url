#!/bin/bash

# 内存监控脚本 - 监控Go应用内存使用情况

echo "🔍 短链接服务内存监控"
echo "监控间隔: 5秒"
echo "按 Ctrl+C 停止监控"
echo

# 创建监控日志文件
LOG_FILE="memory_monitor_$(date +%Y%m%d_%H%M%S).log"

echo "时间,Docker内存(MB),堆内存(MB),系统内存(MB),Goroutine数量,GC次数" > "$LOG_FILE"

while true; do
    TIMESTAMP=$(date '+%H:%M:%S')
    
    # 获取Docker内存使用
    DOCKER_MEM=$(docker stats --no-stream --format "{{.MemUsage}}" shorturl_app 2>/dev/null | cut -d'/' -f1 | sed 's/MiB//' | sed 's/GiB/*1024/' | bc 2>/dev/null)
    
    # 获取Go内存统计（如果接口可用）
    MEMORY_STATS=$(curl -s http://localhost:8080/debug/memory 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$MEMORY_STATS" ]; then
        # 如果内存接口可用，解析详细信息
        HEAP_ALLOC=$(echo "$MEMORY_STATS" | jq -r '.data.before_gc.heap_alloc_mb // "N/A"' 2>/dev/null)
        SYS_MEM=$(echo "$MEMORY_STATS" | jq -r '.data.before_gc.sys_mb // "N/A"' 2>/dev/null)
        GOROUTINES=$(echo "$MEMORY_STATS" | jq -r '.data.before_gc.num_goroutine // "N/A"' 2>/dev/null)
        GC_COUNT=$(echo "$MEMORY_STATS" | jq -r '.data.before_gc.num_gc // "N/A"' 2>/dev/null)
    else
        HEAP_ALLOC="N/A"
        SYS_MEM="N/A"
        GOROUTINES="N/A"
        GC_COUNT="N/A"
    fi
    
    # 显示信息
    printf "%-8s | Docker: %-6s MB | Heap: %-6s MB | Sys: %-6s MB | Goroutines: %-4s | GC: %-4s\n" \
        "$TIMESTAMP" "${DOCKER_MEM:-N/A}" "$HEAP_ALLOC" "$SYS_MEM" "$GOROUTINES" "$GC_COUNT"
    
    # 记录到日志文件
    echo "$TIMESTAMP,${DOCKER_MEM:-0},$HEAP_ALLOC,$SYS_MEM,$GOROUTINES,$GC_COUNT" >> "$LOG_FILE"
    
    sleep 5
done 