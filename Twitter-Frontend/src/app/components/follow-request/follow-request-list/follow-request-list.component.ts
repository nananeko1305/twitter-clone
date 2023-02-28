import { Component, Input } from '@angular/core';
import { FollowRequest } from 'src/app/models/followRequest.model';

@Component({
  selector: 'app-follow-request-list',
  templateUrl: './follow-request-list.component.html',
  styleUrls: ['./follow-request-list.component.css']
})
export class FollowRequestListComponent {

  @Input() followRequests: FollowRequest[] = []

}
