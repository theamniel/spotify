import asyncio
import json
from time import time
from aiohttp import ClientSession, ClientWebSocketResponse, WSMsgType

URL = 'http://localhost:5000/socket'

# opcode 1
async def hello(socket: ClientWebSocketResponse, packet):
  print('[HELLO]:', packet['d'])
  interval = packet['d'].get('heartbeat_interval', 35000) / 1000
  asyncio.create_task(heartbeat_loop(socket, interval)) 
  await socket.send_json({'op': 2})

# opcode 0
def dispatch(packet):
  format = f':{packet.get('t', '')}'
  print(f'[DISPATCH{format}]: {packet['d']}')

last_heartbeat_sent = 0.0
async def heartbeat_loop(socket: ClientWebSocketResponse, interval: float):
  global last_heartbeat_sent
  while True:
    await asyncio.sleep(interval)
    last_heartbeat_sent = time()
    print('[HEARTBEAT]: Sending heartbeat.')
    await socket.send_json({'op': 3})

# opcode 4
def heartbeat_ack():
  last_heartbeat_receive = time()
  result = round((last_heartbeat_receive - last_heartbeat_sent) * 1000, 2)
  print(f'[HEARTBEAT_ACK]: {result}ms')

# opcode 5
def error(packet):
  print(f'[ERROR]:, {packet}')

async def main():
  session = ClientSession()
  async with session.ws_connect(URL) as ws:
    async for msg in ws:
      if msg.type is WSMsgType.TEXT:
        packet = json.loads(msg.data)
        op = packet.get('op', 0)
        
        if op == 1:
          await hello(ws, packet)
        elif op == 0:
          dispatch(packet)
        elif op == 3:
          # heartbeat delay...
          pass
        elif op == 4:
          heartbeat_ack()
        elif op == 5:
          error(packet)
        else:
          print(f'[UNKNOWN]: {packet}')

      if msg.type in (WSMsgType.CLOSED, WSMsgType.ERROR):
        error(msg.data)
        break

if __name__ == '__main__':
  asyncio.run(main())