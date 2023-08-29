import {Component, OnInit} from '@angular/core';
import {AngularFireMessaging} from "@angular/fire/compat/messaging";

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {

  constructor(
    private angularMessaging: AngularFireMessaging,
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
      // Ovde možete upravljati primljenim obaveštenjem
      console.log('Primljeno push obaveštenje:', message);
    });

  }
}


