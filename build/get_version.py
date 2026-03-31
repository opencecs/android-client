#!/usr/bin/env python3
"""读取 build/config.yml 中的版本信息"""

import re
import os
from datetime import datetime, timezone

config_file = os.path.join(os.path.dirname(__file__), '..', 'build', 'config.yml')

try:
    with open(config_file, 'r', encoding='utf-8') as f:
        content = f.read()
    
    match = re.search(r'version:\s*"?([0-9]+\.[0-9]+\.[0-9]+)', content)
    version = match.group(1) if match else "0.0.1"
    
    build_time = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    
    ldflags = f'-X edgeclient/updater.AppVersion={version} -X edgeclient/updater.BuildTime={build_time} -w -s -H windowsgui'
    
    print(f"Version: {version}")
    print(f"BuildTime: {build_time}")
    print(f"LDFLAGS: {ldflags}")
    
    # Taskfile 可以通过环境变量读取这些值
    print(f"##vso[task.setvariable variable=LDFLAGS;]{ldflags}")
    print(f"##vso[task.setvariable variable=APP_VERSION;]{version}")
    
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    exit(1)
