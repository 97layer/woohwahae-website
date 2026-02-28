import os
import re
import datetime

def sync_cache():
    version = datetime.datetime.now().strftime("%Y%m%d%H%M%S")
    target_dirs = ['website', 'website/archive', 'website/practice', 'website/about', 'website/lab']
    
    print(f"🔄 Syncing Asset Cache Versions to: {version}")
    updated_count = 0

    # Pattern to match original or corrupted style/js
    # Fix: Remove literal / to match both relative and absolute paths
    # Corrupted: href="...assets/css/P..." (Where P is from octal \120)
    style_pattern = re.compile(r'(href="[^"]*assets/css/)(?:style\.css\?v=[a-zA-Z0-9]+|P[0-9]+)')
    js_pattern = re.compile(r'(src="[^"]*assets/js/)(?:site\.js\?v=[a-zA-Z0-9]+|P[0-9]+)')

    for root_dir in target_dirs:
        if not os.path.exists(root_dir): continue
        for filename in os.listdir(root_dir):
            if filename.endswith('.html'):
                filepath = os.path.join(root_dir, filename)
                try:
                    with open(filepath, 'r', encoding='utf-8') as f:
                        content = f.read()

                    # Fix style.css - USE \g<1> to avoid octal interpretation
                    new_content = style_pattern.sub(fr'\g<1>style.css?v={version}', content)
                    new_content = js_pattern.sub(fr'\g<1>site.js?v={version}', new_content)

                    if new_content != content:
                        with open(filepath, 'w', encoding='utf-8') as f:
                            f.write(new_content)
                        print(f"  ✅ Recovered & Updated: {os.path.relpath(filepath, 'website')}")
                        updated_count += 1
                except Exception as e:
                    print(f"  ❌ Error reading {filepath}: {e}")

    print(f"✨ Successfully updated and RECOVERED cache-bust version in {updated_count} files.")

if __name__ == '__main__':
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    os.chdir(project_root)
    sync_cache()
