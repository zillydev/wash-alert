self.addEventListener('push', (event) => {
  const data = event.data ? event.data.text() : null;

  // Handle the push message here
  console.log('Push message received', data);

  event.waitUntil(self.registration.showNotification('Push Notification', {
    body: data,
  }));
});

self.addEventListener('notificationclick', (event) => {
  // Handle notification click, such as opening the app or performing an action
  event.notification.close();
  event.waitUntil(
    clients.openWindow('/')  // This could open the app or take other actions
  );
});

self.addEventListener('activate', (event) => {
  console.log('Service Worker activated');
  event.waitUntil(self.clients.claim());
});
