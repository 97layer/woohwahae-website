import os
import re

def inject_components():
    component_dir = 'website/_components'
    target_dirs = ['website', 'website/archive', 'website/practice', 'website/about', 'website/lab']
    
    components = {}
    for filename in os.listdir(component_dir):
        if filename.endswith('.html'):
            name = filename[:-5]
            with open(os.path.join(component_dir, filename), 'r', encoding='utf-8') as f:
                content = f.read().strip()
                components[name] = content

    print(f"🚀 Starting Global Component Injection...")
    updated_count = 0

    for root_dir in target_dirs:
        if not os.path.exists(root_dir): continue
        for filename in os.listdir(root_dir):
            if filename.endswith('.html'):
                filepath = os.path.join(root_dir, filename)
                with open(filepath, 'r', encoding='utf-8') as f:
                    content = f.read()

                new_content = content
                for name, comp_content in components.items():
                    pattern = rf'<!-- COMPONENT:{name} -->.*?<!-- /COMPONENT:{name} -->'
                    replacement = f'<!-- COMPONENT:{name} -->\n{comp_content}\n<!-- /COMPONENT:{name} -->'
                    
                    if re.search(pattern, new_content, re.DOTALL):
                        # print(f"  [UP] {name} in {os.path.relpath(filepath, 'website')}")
                        new_content = re.sub(pattern, replacement, new_content, flags=re.DOTALL)
                    else:
                        # If not found but in components, we don't automatically inject to avoid breaking layouts
                        # except for basic ones if needed.
                        pass

                if new_content != content:
                    with open(filepath, 'w', encoding='utf-8') as f:
                        f.write(new_content)
                    print(f"  [UP] components in {os.path.relpath(filepath, 'website')}")
                    updated_count += 1

    print(f"\n✨ Successfully updated components in {updated_count} files.")

if __name__ == '__main__':
    # Ensure we are in the project root
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    os.chdir(project_root)
    inject_components()
