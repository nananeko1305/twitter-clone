import {Component, OnInit} from '@angular/core';
import {AngularFireMessaging} from "@angular/fire/compat/messaging";


@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {

  constructor(
    private angularMessaging: AngularFireMessaging
  ) {}

  ngOnInit() {

    const requestPermission = this.angularMessaging.requestPermission
    requestPermission.subscribe(response => {
      console.log(response)
    })

    const requestToken = this.angularMessaging.getToken
    requestToken.subscribe(response => {
      console.log(response)
    })

  }
}


