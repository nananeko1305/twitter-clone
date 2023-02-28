import { Component, EventEmitter, OnInit, Output } from '@angular/core';
import { FollowRequest } from 'src/app/models/followRequest.model';
import { User } from 'src/app/models/user.model';
import { FollowService } from 'src/app/services/follow.service';
import { UserService } from 'src/app/services/user.service';

@Component({
  selector: 'app-follow-requests',
  templateUrl: './follow-requests.component.html',
  styleUrls: ['./follow-requests.component.css']
})
export class FollowRequestsComponent implements OnInit {

  @Output() requestAnswer = new EventEmitter<string>();

  constructor(private userService: UserService,
              private followService: FollowService) { }

  user: User = new User();
  requests: FollowRequest[] = []
  firstFollowRequest: FollowRequest = new FollowRequest();
  secondFollowRequest: FollowRequest = new FollowRequest();

  followRequests: FollowRequest[] = []

  ngOnInit(): void {
    this.userService.GetMe()
      .subscribe({
        next: (data: User) => {
          this.user = data;
        },
        error: (error) => {
          console.log(error);
        } 
      });

    this.followService.GetRequestsForUser().
      subscribe({
        next: (data) => {
          this.requests = data;         
        },
        error: (error) => {
          console.log(error);
        }
      });
  }


AcceptRequest(id: string){
  this.followService.AcceptRequest(id)
    .subscribe({
      next: (data) => {
        alert("Request Accepted");
      },
      error: (error) => {
        console.log(error);
      },
      complete: () => {
        this.followService.GetRequestsForUser().
        subscribe({
          next: (data) => {
            this.requests = data;         
          },
          error: (error) => {
            console.log(error);
          }
        });
      }
    }
  )
}

DeclineRequest(id: string){
  this.followService.DeclineRequest(id)
    .subscribe({
      next: (data) => {
        alert("Request Denied");
      },
      error: (error) => {
        console.log(error);
      },
      complete: () => {
        this.followService.GetRequestsForUser().
        subscribe({
          next: (data) => {
            this.requests = data;         
          },
          error: (error) => {
            console.log(error);
          }
        });
      }
    }
  )
}



}
