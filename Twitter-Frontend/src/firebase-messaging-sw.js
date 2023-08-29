importScripts('https://www.gstatic.com/firebasejs/8.10.1/firebase-app.js');
importScripts('https://www.gstatic.com/firebasejs/8.10.1/firebase-messaging.js');

firebase.initializeApp({
  apiKey: "AIzaSyDjxiCAz3Q0p1RIUUMLJoEEnZA5Cs407og",
  authDomain: "twitterclone-1a08f.firebaseapp.com",
  projectId: "twitterclone-1a08f",
  storageBucket: "twitterclone-1a08f.appspot.com",
  messagingSenderId: "142880253564",
  appId: "1:142880253564:web:76e07e1a02e286d3a0cbf9"
});


const messaging = firebase.messaging();

messaging.onBackgroundMessage((payload) => {
  console.log(
    '[firebase-messaging-sw.js] Received background message ',
    payload
  );
  // Customize notification here
  const notificationTitle = 'Test';
  const notificationOptions = {
    body: 'Test',
  };

  self.registration.showNotification(notificationTitle, notificationOptions)
});


console.log("Script started")
