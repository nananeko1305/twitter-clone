import {Component, OnInit} from '@angular/core';
import {AngularFireMessaging} from "@angular/fire/compat/messaging";
import { SwPush } from '@angular/service-worker';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {

  constructor(
    private angularMessaging: AngularFireMessaging,
    private swPush: SwPush
  ) {}

  ngOnInit() {

    this.angularMessaging.requestPermission.subscribe(response => {
      console.log(response)
    })

    const requestToken = this.angularMessaging.getToken
    requestToken.subscribe(response => {
      if(response != null)
      localStorage.setItem("fcmToken", response)
    })

    this.angularMessaging.messages.subscribe(message => {

      if (!this.swPush.isEnabled) {
        Notification.requestPermission().then(permission => {
          if (permission === 'granted') {
            const notification = new Notification('Primljeno obaveštenje', {
              body: message.notification?.body
            });

            notification.onclick = event => {
              event.preventDefault();
              // Ovde možete upravljati akcijom kada korisnik klikne na notifikaciju
            };
          }
        });
      }

    });
  }
}


