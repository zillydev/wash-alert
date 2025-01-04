<script setup lang="ts">
import { onMounted, provide, ref } from 'vue';
import { v4 as uuidv4 } from 'uuid';
import { openDB } from 'idb'
import { send } from './utils/websocket';

const socket: WebSocket = new WebSocket("ws://localhost:8080/ws");
provide('socket', socket);
const tokenRef = ref('')
provide('token', tokenRef)

onMounted(async () => {
  const db = await openDB('tokenDb', 1, {
    upgrade(db) {
      db.createObjectStore('tokenStore', { keyPath: 'id' });
    }
  });

  const tokenStore = db.transaction('tokenStore', 'readwrite').objectStore('tokenStore');
  var token = await tokenStore.get('token');
  console.log(token)

  if (!token) {
    token = uuidv4();
    await tokenStore.add({ id: 'token', value: token });
  } else {
    token = token.value
  }

  tokenRef.value = token

  if ('serviceWorker' in navigator) {
    try {
      const registration = await navigator.serviceWorker.register('/service-worker.js');
      console.log('Service Worker registered with scope:', registration.scope);

      // Request push notification permission
      const permission = await Notification.requestPermission();
      if (permission === 'granted') {
        console.log('Notification permission granted');
        // You can now subscribe for push notifications
        await subscribeUserToPush(registration, token);
      }
    } catch (error) {
      console.error('Service Worker registration failed:', error);
    }
  }
})

async function subscribeUserToPush(registration: ServiceWorkerRegistration, token: string) {
  try {
    const subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: 'BL86hEFii0qrpPpeVfk0ulYtSdsV91jPfxhqcvrVth7Olum3B6w3IB4H6VyVxOovjjhtkaRFNuKpm4kISNR5oSw',
    });
    console.log('User is subscribed to push notifications', subscription);
    // You can send the subscription to your server to trigger push notifications
    send(socket, 'subscribe', { "Token": token, "Subscription": subscription })
  } catch (error) {
    console.error('Failed to subscribe to push notifications', error);
  }
}
</script>

<template>
  <Suspense>
    <template #default>
      <router-view></router-view>
    </template>
    <template #fallback>
      <div>Loading...</div>
    </template>
  </Suspense>
</template>
