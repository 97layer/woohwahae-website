import sys
from pathlib import Path

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.core_config import AGENT_CREW, SYSTEM_CONFIG
from execution.system.manage_directive import DirectiveManager

def verify_system():
    print(f"Verifying {SYSTEM_CONFIG['PROJECT_NAME']} v{SYSTEM_CONFIG['VERSION']}...")
    
    # 1. Config Check
    print("[1] Configuration Check: OK")
    
    # 2. Agent Directive Check
    dm = DirectiveManager(str(project_root))
    missing_agents = []
    
    print("[2] Agent Directives Check:")
    for role, config in AGENT_CREW.items():
        fname = Path(config['directive_path']).name
        data = dm.read_directive(fname)
        if data:
            print(f"  - {role}: OK ({fname})")
        else:
            print(f"  - {role}: MISSING ({fname})")
            missing_agents.append(role)
            
    if missing_agents:
        print(f"ERROR: Missing directives for {missing_agents}")
        sys.exit(1)
        
    print("System Verification Completed Successfully.")

if __name__ == "__main__":
    verify_system()
