import os
import re
from bs4 import BeautifulSoup

ROOT = os.path.abspath('website')

def get_depth_prefix(filepath):
    # Calculate depth relative to ROOT
    rel_path = os.path.relpath(filepath, ROOT)
    depth = rel_path.count(os.sep)
    return "../" * depth if depth > 0 else "./"

def get_unified_nav(prefix):
    # Path adjusting for root links
    return f'''
  <nav class="site-nav" id="site-nav">
    <a href="{prefix}" class="nav-brand" aria-label="WOOHWAHAE">
      <img src="{prefix}assets/img/symbol.png" alt="WOOHWAHAE" class="nav-symbol">
    </a>
    <ul class="nav-links" id="nav-links">
      <li><a href="{prefix}archive/">Archive</a></li>
      <li><a href="{prefix}practice/">Practice</a></li>
      <li><a href="{prefix}about/">About</a></li>
    </ul>
    <button class="nav-toggle" id="nav-toggle" aria-label="메뉴" aria-expanded="false">
      <span></span><span></span>
    </button>
  </nav>

  <div class="nav-overlay" id="nav-overlay">
    <a href="{prefix}archive/">Archive</a>
    <a href="{prefix}practice/">Practice</a>
    <a href="{prefix}about/">About</a>
  </div>
'''

def get_unified_wave_bg():
    return '''
  <!-- wave-bg 추가 -->
  <div class="wave-bg" aria-hidden="true"
    style="position:fixed;top:50%;left:50%;
           transform:translate(-50%,-50%);
           width:120vmax;height:120vmax;
           z-index:0;pointer-events:none;opacity:0.6">
    <svg viewBox="0 0 800 800" width="100%" height="100%">
      <g fill="none" stroke="#D5D4CF" stroke-width="0.5" opacity="0.4">
        <circle cx="400" cy="400" r="60" opacity="0.4">
          <animate attributeName="r" values="60;70;60" dur="3.8s" begin="0s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.4;0.2;0.4" dur="3.8s" begin="0s" repeatCount="indefinite" />
        </circle>
        <circle cx="400" cy="400" r="140" opacity="0.3">
          <animate attributeName="r" values="140;155;140" dur="3.8s" begin="0.4s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.3;0.15;0.3" dur="3.8s" begin="0.4s" repeatCount="indefinite" />
        </circle>
        <circle cx="400" cy="400" r="240" opacity="0.2">
          <animate attributeName="r" values="240;260;240" dur="3.8s" begin="0.8s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.2;0.1;0.2" dur="3.8s" begin="0.8s" repeatCount="indefinite" />
        </circle>
        <circle cx="400" cy="400" r="360" opacity="0.12">
          <animate attributeName="r" values="360;385;360" dur="3.8s" begin="1.2s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.12;0.05;0.12" dur="3.8s" begin="1.2s" repeatCount="indefinite" />
        </circle>
      </g>
    </svg>
  </div>
'''

def get_unified_footer(prefix):
    return f'''
  <footer class="site-footer">
    <div class="footer-inner">
      <div class="footer-contact">
        <span class="footer-contact__label">Contact</span>
        <div class="footer-contact__grid">
          <a href="mailto:hello@woohwahae.kr">hello@woohwahae.kr</a>
          <a href="https://instagram.com/woohwahae" target="_blank" rel="noopener">@woohwahae</a>
        </div>
      </div>
      <div class="footer-bottom">
        <nav class="footer-nav">
          <a href="{prefix}archive/">Archive</a>
          <a href="{prefix}practice/">Practice</a>
          <a href="{prefix}about/">About</a>
        </nav>
        <div class="footer-legal">
          <a href="{prefix}privacy.html">Privacy</a>
          <a href="{prefix}terms.html">Terms</a>
        </div>
        <p class="footer-copy">&copy; 2026 WOOHWAHAE</p>
      </div>
    </div>
  </footer>
'''

def process_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        prefix = get_depth_prefix(filepath)

        # Parse with BeautifulSoup using 'html.parser' (it tries to preserve structure)
        soup = BeautifulSoup(content, 'html.parser')
        body = soup.find('body')

        if not body:
            print(f"Skipped (No body tag): {filepath}")
            return

        # 1. Remove existing nav, nav-overlay, wave-bg, and footer directly descending from body
        # We need to collect elements to decompose first
        to_decompose = []
        for child in body.children:
            if getattr(child, 'name', None) == 'nav':
                to_decompose.append(child)
            elif getattr(child, 'name', None) == 'div' and child.get('class'):
                if 'nav-overlay' in child.get('class') or 'wave-bg' in child.get('class'):
                    to_decompose.append(child)
            elif getattr(child, 'name', None) == 'footer':
                to_decompose.append(child)

        for elem in to_decompose:
            elem.decompose()

        # 2. Prepare new unified elements
        nav_html = get_unified_nav(prefix)
        nav_soup = BeautifulSoup(nav_html, 'html.parser')

        wave_html = get_unified_wave_bg()
        wave_soup = BeautifulSoup(wave_html, 'html.parser')

        footer_html = get_unified_footer(prefix)
        footer_soup = BeautifulSoup(footer_html, 'html.parser')

        # Insert Nav (prepended to body)
        # Using reversed to maintain parsed order when inserting at 0
        nodes_to_insert = [node for node in nav_soup.contents if node.name or str(node).strip()]
        for node in reversed(nodes_to_insert):
            body.insert(0, node)
        
        # Insert Wave Background
        # Right after nav elements if possible, or just prepend as well
        overlay = body.find('div', class_='nav-overlay')
        wave_nodes = [node for node in wave_soup.contents if node.name or str(node).strip()]
        if overlay:
            for node in reversed(wave_nodes):
                overlay.insert_after(node)
        else:
            for node in reversed(wave_nodes):
                body.insert(0, node)

        # Insert Footer (append to body)
        footer_nodes = [node for node in footer_soup.contents if node.name or str(node).strip()]
        for node in footer_nodes:
            body.append(node)

        # 3. Overwrite
        # Using html.parser, the str(soup) output is usually good
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(str(soup))
            
        print(f"Standardized: {filepath}")
    except PermissionError:
        print(f"Skipped (Permission Denied): {filepath}")
    except Exception as e:
        print(f"Error processing {filepath}: {e}")

def main():
    skip_dirs = ['_templates', 'assets', 'api', '.git']
    for root, dirs, files in os.walk(ROOT):
        dirs[:] = [d for d in dirs if d not in skip_dirs]
        for file in files:
            if file.endswith('.html'):
                filepath = os.path.join(root, file)
                process_file(filepath)
                
if __name__ == '__main__':
    main()
