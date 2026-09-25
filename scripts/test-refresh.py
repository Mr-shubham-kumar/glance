"""Development-only CDP smoke test: manual refresh restores the selected Explore tab.

Requires Chrome, requests and websocket-client locally. Run with Glance on :8080.
"""
import json
import subprocess
import tempfile
import time
from pathlib import Path
import requests
import websocket

chrome = Path('C:/Program Files/Google/Chrome/Application/chrome.exe')
profile = tempfile.TemporaryDirectory()
process = subprocess.Popen([str(chrome), '--headless', '--disable-gpu', '--no-first-run', '--remote-allow-origins=*', '--remote-debugging-port=9231', f'--user-data-dir={profile.name}', 'about:blank'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(60):
        try:
            pages = requests.get('http://127.0.0.1:9231/json', timeout=1).json()
            break
        except requests.RequestException:
            time.sleep(.2)
    ws = websocket.create_connection(next(p['webSocketDebuggerUrl'] for p in pages if p['type']=='page'),timeout=90)
    serial = 0
    def call(method, params=None):
        global serial
        serial += 1
        ws.send(json.dumps({'id':serial,'method':method,'params':params or {}}))
        while True:
            message = json.loads(ws.recv())
            if message.get('id') == serial:
                if 'error' in message: raise RuntimeError(message['error'])
                return message.get('result',{})
    def eval_js(expression):
        return call('Runtime.evaluate',{'expression':expression,'returnByValue':True})['result'].get('value')
    call('Page.navigate',{'url':'http://127.0.0.1:8080/explore'})
    time.sleep(12)
    result = eval_js("document.querySelectorAll('.widget-type-group .widget-group-title').length")
    if result < 3: raise RuntimeError(f'Explore tabs missing: {result}')
    eval_js("document.querySelectorAll('.widget-type-group .widget-group-title')[2].click(); window.scrollTo(0, 350); document.querySelectorAll('.page-refresh')[0].click(); true")
    time.sleep(12)
    selected = eval_js("document.querySelector('.widget-type-group .widget-group-title-current')?.textContent.trim()")
    state = eval_js("sessionStorage.getItem('signal-desk:refresh:explore')")
    print('Restored tab:',selected,'Temporary session state cleared:',state is None)
    if selected != 'Hardware & systems' or state is not None: raise RuntimeError('refresh did not restore the selected tab')
    eval_js("const actualNow = Date.now; Date.now = () => actualNow() + 11 * 60 * 1000; window.dispatchEvent(new Event('focus')); true")
    time.sleep(12)
    selected = eval_js("document.querySelector('.widget-type-group .widget-group-title-current')?.textContent.trim()")
    state = eval_js("sessionStorage.getItem('signal-desk:refresh:explore')")
    print('Return-after-ten-minutes refresh restored tab:', selected, 'State cleared:', state is None)
    if selected != 'Hardware & systems' or state is not None: raise RuntimeError('return-to-tab refresh failed')
    ws.close()
finally:
    process.terminate()
    try:
        process.wait(timeout=5)
        profile.cleanup()
    except (OSError,subprocess.TimeoutExpired):
        pass
