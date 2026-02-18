"""
캐시 비활성화 개발 서버
— 모든 응답에 no-cache 헤더를 붙여 브라우저 캐시 문제를 방지한다.
— 사용법: python serve.py [port]
"""
import http.server
import sys

class NoCacheHandler(http.server.SimpleHTTPRequestHandler):
    """모든 응답에 캐시 비활성화 헤더를 추가하는 핸들러"""

    def end_headers(self):
        self.send_header('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
        self.send_header('Pragma', 'no-cache')
        self.send_header('Expires', '0')
        super().end_headers()

if __name__ == '__main__':
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8081
    server = http.server.HTTPServer(('', port), NoCacheHandler)
    print(f'개발 서버 시작: http://localhost:{port} (캐시 비활성화)')
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print('\n서버 종료')
