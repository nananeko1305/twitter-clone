import { Component, Input, Output, EventEmitter } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { FollowRequest } from 'src/app/models/followRequest.model';
import { FollowService } from 'src/app/services/follow.service';

@Component({
  selector: 'app-follow-request-item',
  templateUrl: './follow-request-item.component.html',
  styleUrls: ['./follow-request-item.component.css']
})
export class FollowRequestItemComponent {

  @Input() followRequest: FollowRequest = new FollowRequest();
  @Output() answerFollowRequest = new EventEmitter<any>()

  constructor(private followService: FollowService,
    private _snackBar: MatSnackBar) { }

  AcceptRequest(id: string){
    this.followService.AcceptRequest(id)
      .subscribe({
        next: (data) => {
          this.answerFollowRequest.emit();
          this.openSnackBar("Request Accepted","")
        },
        error: (error) => {
          console.log(error);
        }
      }
    )
  }
  
  DeclineRequest(id: string){
    this.followService.DeclineRequest(id)
      .subscribe({
        next: (data) => {
          this.openSnackBar("Request Denied","")
        },
        error: (error) => {
          console.log(error);
        }
      }
    )
  }

  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action,  {
      duration: 3000
    });
  }

}
