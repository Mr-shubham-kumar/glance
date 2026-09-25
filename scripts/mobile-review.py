"""Capture an actual mobile-emulated screenshot using local Chrome's CDP.

Requires websocket-client and requests in the development Python environment.
Chrome must be installed; no dependency is added to the deployed container.
"""
import base64
import json
import subprocess
import sys
import tempfile
import time
from pathlib import Path
import requests
import websocket

chrome = Path('C:/Program Files/Google/Chrome/Application/chrome.exe')
profile = tempfile.TemporaryDirectory()
process = subprocess.Popen([str(chrome), '--headless', '--disable-gpu', '--no-first-run', '--remote-allow-origins=*', '--remote-debugging-port=9229', f'--user-data-dir={profile.name}', 'about:blank'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(50):
        try:
            pages = requests.get('http://127.0.0.1:9229/json', timeout=1).json()
            break
        except requests.RequestException:
            time.sleep(.2)
    ws = websocket.create_connection(next(p['webSocketDebuggerUrl'] for p in pages if p['type'] == 'page'), timeout=40)
    sequence = 0
    def call(method, params=None):
        global sequence
        sequence += 1
        ws.send(json.dumps({'id': sequence, 'method': method, 'params': params or {}}))
        while True:
            event = json.loads(ws.recv())
            if event.get('id') == sequence:
                if 'error' in event: raise RuntimeError(event['error'])
                return event.get('result', {})
    call('Emulation.setDeviceMetricsOverride', {'width':390, 'height':844, 'deviceScaleFactor':1, 'mobile':True})
    call('Emulation.setTouchEmulationEnabled', {'enabled':True})
    call('Page.navigate', {'url':sys.argv[1] if len(sys.argv)>1 else 'https://glance-9gmu.onrender.com/today'})
    time.sleep(15)
    metrics = call('Runtime.evaluate', {'expression':'JSON.stringify({width:innerWidth,body:document.body.scrollWidth,root:document.documentElement.scrollWidth})'})
    print('Viewport', metrics['result']['value'])
    screenshot = call('Page.captureScreenshot', {'format':'png'})
    Path('mobile-cdp.png').write_bytes(base64.b64decode(screenshot['data']))
    ws.close()
finally:
    process.terminate()
    try:
        process.wait(timeout=5)
        profile.cleanup()
    except (OSError, subprocess.TimeoutExpired):
        # Chrome may leave profile cache files open briefly on Windows.
        pass
