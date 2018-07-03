'use strict';

self.addEventListener('push', function(event) {
  console.log('Received a push message', event);
  var notify = event.data.json();
  event.waitUntil(
    self.registration.showNotification(notify.title, notify)
  );
});

self.addEventListener('notificationclick', function(event) {
  console.log('On notification click: ', event.notification.tag);
  // Android doesn’t close the notification when you click on it
  // See: http://crbug.com/463146
  event.notification.close();
  var data = event.notification.data;
  // This looks to see if the current is already open and
  // focuses if it is
  event.waitUntil(clients.openWindow(data.url));
});
