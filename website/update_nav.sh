#!/bin/bash

# 7개 세션 네비게이션 HTML
NAV_HTML='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
    </ul>'

# 푸터 네비게이션 HTML
FOOTER_NAV='        <ul class="footer-nav-list">
          <li><a href="about.html">About</a></li>
          <li><a href="archive/">Archive</a></li>
          <li><a href="shop.html">Shop</a></li>
          <li><a href="atelier.html">Atelier</a></li>
          <li><a href="playlist.html">Playlist</a></li>
          <li><a href="project.html">Project</a></li>
          <li><a href="photography.html">Photography</a></li>
        </ul>'

# 각 페이지별 active 클래스 설정
declare -A ACTIVE_CLASSES=(
    ["shop.html"]='<li><a href="shop.html" class="active">Shop</a></li>'
    ["atelier.html"]='<li><a href="atelier.html" class="active">Atelier</a></li>'
    ["playlist.html"]='<li><a href="playlist.html" class="active">Playlist</a></li>'
    ["project.html"]='<li><a href="project.html" class="active">Project</a></li>'
    ["photography.html"]='<li><a href="photography.html" class="active">Photography</a></li>'
    ["contact.html"]='<li><a href="contact.html" class="active">Contact</a></li>'
)

# 각 페이지 수정
for file in shop.html atelier.html playlist.html project.html photography.html contact.html; do
    echo "Updating $file..."

    # Create temp navigation with active class
    TEMP_NAV="$NAV_HTML"
    if [[ -n "${ACTIVE_CLASSES[$file]}" ]]; then
        # Replace the corresponding line with active class
        case $file in
            shop.html)
                TEMP_NAV='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html" class="active">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
    </ul>'
                ;;
            atelier.html)
                TEMP_NAV='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html" class="active">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
    </ul>'
                ;;
            playlist.html)
                TEMP_NAV='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html" class="active">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
    </ul>'
                ;;
            project.html)
                TEMP_NAV='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html" class="active">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
    </ul>'
                ;;
            photography.html)
                TEMP_NAV='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html" class="active">Photography</a></li>
    </ul>'
                ;;
            contact.html)
                TEMP_NAV='    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
      <li><a href="contact.html" class="active">Contact</a></li>
    </ul>'
                ;;
        esac
    fi

    # Create temporary file
    cp "$file" "$file.bak"

    # Use perl to replace navigation (handles multi-line better than sed)
    perl -i -0pe 's/<ul class="nav-links">.*?<\/ul>/$ENV{TEMP_NAV}/s' "$file"

    # Update footer navigation
    perl -i -0pe 's/<ul class="footer-nav-list">.*?<\/ul>/$ENV{FOOTER_NAV}/s' "$file"

    export TEMP_NAV FOOTER_NAV
done

# Update index.html footer too
echo "Updating index.html footer..."
perl -i -0pe 's/<ul class="footer-nav-list">.*?<\/ul>/$ENV{FOOTER_NAV}/s' index.html

echo "Navigation update complete!"