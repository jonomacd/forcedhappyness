var isPushEnabled = false;

function unsubscribe() {
    navigator.serviceWorker.ready.then(function(serviceWorkerRegistration) {
      // To unsubscribe from push messaging, you need get the
      // subcription object, which you can call unsubscribe() on.
      serviceWorkerRegistration.pushManager.getSubscription().then(
        function(pushSubscription) {
          // Check we have a subscription to unsubscribe
          if (!pushSubscription) {
            // No subscription object, so set the state
            // to allow the user to subscribe to push
            isPushEnabled = false;          
            return;
          }
  
          // TODO: Make a request to your server to remove
          // the users data from your data store so you
          // don't attempt to send them push messages anymore
  
          // We have a subcription, so call unsubscribe on it
          pushSubscription.unsubscribe().then(function() {
            
            isPushEnabled = false;
            window.location.href = '/notifications';
          }).catch(function(e) {
            // We failed to unsubscribe, this can lead to
            // an unusual state, so may be best to remove
            // the subscription id from your data store and
            // inform the user that you disabled push
  
            console.log('Unsubscription error: ', e);
            
          });
        }).catch(function(e) {
          console.log('Error thrown while unsubscribing from ' +
            'push messaging.', e);
        });
    });
}
  
  function subscribe(applicationServerKey) {
    if (isPushEnabled) {
        // Push is already enabled. Why are we calling subscribe????
      return;
    }

    navigator.serviceWorker.ready.then(function(serviceWorkerRegistration) {
      serviceWorkerRegistration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(applicationServerKey)
      })
        .then(function(subscription) {
          // The subscription was successful
          isPushEnabled = true;        
            
          document.getElementById('hidden-sub').value =  JSON.stringify(subscription);
          document.getElementById('hidden-notification-form').submit();          
          return true
        })
        .catch(function(e) {
          if (Notification.permission === 'denied') {
            // The user denied the notification permission which
            // means we failed to subscribe and the user will need
            // to manually change the notification permission to
            // subscribe to push messages
            var snackbarContainer = document.querySelector('#error-snackbar');
            var data = {message: 'You have denied permissions for notifications'};
            snackbarContainer.MaterialSnackbar.showSnackbar(data);
            console.log('Permission for Notifications was denied');          
          } else {
            // A problem occurred with the subscription, this can
            // often be down to an issue or lack of the gcm_sender_id
            // and / or gcm_user_visible_only
            var snackbarContainer = document.querySelector('#error-snackbar');
            var data = {message: 'hmm... something went wrong'};
            snackbarContainer.MaterialSnackbar.showSnackbar(data);
            console.log('Unable to subscribe to push.', e);          
          }
        });
    });
  }
  
  // Once the service worker is registered set the initial state
  function initialiseState() {
    // Are Notifications supported in the service worker?
    if (!('showNotification' in ServiceWorkerRegistration.prototype)) {
      console.log('Notifications aren\'t supported.');
      document.getElementById('enable-notifications').style.display = 'none';
      document.getElementById('your-device-sucks').style.display = 'block';
      return;
    }
  
    // Check the current Notification permission.
    // If its denied, it's a permanent block until the
    // user changes the permission
    if (Notification.permission === 'denied') {
      console.log('The user has blocked notifications.');
      document.getElementById('notification-text').innerText = 'Notifications have been blocked. Setup again to reenable';
      document.getElementById('enable-notifications').style.display = 'block';
      return;
    }
  
    // Check if push messaging is supported
    if (!('PushManager' in window)) {
      console.log('Push messaging isn\'t supported.');
      document.getElementById('enable-notifications').style.display = 'none';
      document.getElementById('your-device-sucks').style.display = 'block';
      return;
    }
  
    // We need the service worker registration to check for a subscription
    return navigator.serviceWorker.ready.then(function(serviceWorkerRegistration) {
      // Do we already have a push message subscription?
      return serviceWorkerRegistration.pushManager.getSubscription()
        .then(function(subscription) {
          
          if (!subscription) {
            document.getElementById('enable-notifications').style.display = 'block';
            return;
          }

          document.getElementById('enable-notifications').style.display = 'none';
          document.getElementById('enabled-notifications').style.display = 'block';
          // Keep your server in sync with the latest subscription
          //sendSubscriptionToServer(subscription);
  
          // Set your UI to show they have subscribed for
          // push messages        
          isPushEnabled = true;
        })
        .catch(function(err) {
          console.log('Error during getSubscription()', err);
        });
    }).catch(function(err) {
        console.log('Error during serviceworker.ready()', err);
    });
  }
  
  window.addEventListener('load', function() {
    
    // Check that service workers are supported, if so, progressively
    // enhance and add push messaging support, otherwise continue without it.
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('/service-worker.js')
      .then(function(){      
        initialiseState()
          .then(function(){        
            console.log("State inited");
          })});
    } else {
      console.log('Service workers aren\'t supported in this browser.');
      document.getElementById('enable-notifications').style.display = 'none';
      document.getElementById('your-device-sucks').style.display = 'block';
    }
  });
  
  function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding)
      .replace(/\-/g, '+')
      .replace(/_/g, '/')
    ;
    const rawData = window.atob(base64);
    return Uint8Array.from([...rawData].map((char) => char.charCodeAt(0)));
  }