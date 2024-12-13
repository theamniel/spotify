import WebSocket from 'ws';

const ws = new WebSocket('ws://localhost:5050/socket');

let lastHeartbeatReceive: number | null;
let lastHeartbeatSent: number | null;
let interval: number;
let heartbeatInterval: NodeJS.Timeout;
let heartbeat = (v: number) => setInterval(() => {
  lastHeartbeatSent = Date.now();
  ws.send('{"op": 3}');
}, v);

ws.on('open', () => {
  console.log('Connected to WebSocket Server...');
});

ws.on('message', (m) => {
  let packet = JSON.parse(m.toString('utf-8'));
  if (packet.op === 1) {
    console.log('[HELLO]: ', packet.d);
    interval = packet.d.heartbeat_interval;
    heartbeatInterval = heartbeat(packet.d.heartbeat_interval);
    ws.send('{"op":2}');

  } else if (packet.op === 0) {
    console.log(`[DISPATCH${packet.t ? ':' + packet.t : ''}]: `, packet.d);
  
  } else if (packet.op === 3) {
    console.log('[HEARTBEAT]: Abnormal heartbeat...');
    
  } else if (packet.op === 4) {
    lastHeartbeatReceive = Date.now();
    console.log(`[LATENCY]: ${lastHeartbeatReceive - (lastHeartbeatSent ?? 0)}ms`);

  } else {
    console.log('[UNKNOWN]:', packet);
  }
});

ws.on('close', (code, b) => {
  console.log('Connection Closed', code, b.toString());
  clearInterval(interval);
});

ws.on('error', console.error);