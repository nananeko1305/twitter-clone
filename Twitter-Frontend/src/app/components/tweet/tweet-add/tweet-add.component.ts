import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AddTweetDTO } from 'src/app/dto/addTweetDTO';
import { AdConfig } from 'src/app/models/adConfig';
import { Tweet } from 'src/app/models/tweet.model';
import { User } from 'src/app/models/user.model';
import { FollowService } from 'src/app/services/follow.service';
import { TweetService } from 'src/app/services/tweet.service';
import { UserService } from 'src/app/services/user.service';


@Component({
  selector: 'app-tweet-add',
  templateUrl: './tweet-add.component.html',
  styleUrls: ['./tweet-add.component.css']
})
export class TweetAddComponent implements OnInit {

  constructor(
    private formBuilder: FormBuilder,
    private tweetService: TweetService,
    private userService: UserService,
    private router: Router,
    private http: HttpClient,
    private followService: FollowService
  ) 
  { }

  tweetFormGroup: FormGroup = new FormGroup({
    text: new FormControl(''),
    image: new FormControl('')
  });

  advertisementFormGroup: FormGroup = new FormGroup({
    residence: new FormControl(''),
    gender: new FormControl(''),
    age_from: new FormControl(''),
    age_to: new FormControl('')
  })

  file!: File;
  formData = new FormData();

  isChecked = false;
  submittedTweet = false;
  submittedAdvertisement = false;
  user: User = new User();

  ngOnInit(): void {
    this.tweetFormGroup = this.formBuilder.group({
      text: ['', [Validators.required, Validators.minLength(5), Validators.maxLength(280)]], // Validators.pattern('[-_a-zA-Z0-9]*')
      image: ['']
    })

    this.advertisementFormGroup = this.formBuilder.group({
      residence: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(35)]],
      gender: ['', [Validators.required]],
      age_from: ['', [Validators.required, Validators.min(18), Validators.max(100)]],
      age_to: ['', [Validators.required, Validators.min(18), Validators.max(100)]]
    })

    this.userService.GetMe()
      .subscribe({
        next: (data: User) => {
            this.user = data;
        },
        error: (error) => {
          console.log(error);
        }
      })
  }

  isBusiness(): boolean {
    if (this.user.userType == "Business") {
      return true;
    } else {
      return false
    }
  }

  get tweetForm(): { [key: string]: AbstractControl } {
    return this.tweetFormGroup.controls;
  }

  get advertisementForm(): { [key: string]: AbstractControl } {
    return this.advertisementFormGroup.controls;
  }

  check() {
    if (this.isChecked == true) {
      this.isChecked = false
    } else {
      this.isChecked = true
    }
  }

  onSubmit() {
    this.submittedTweet = true;
    this.submittedAdvertisement = true;

    if (this.tweetFormGroup.invalid) {
      return;
    }

    let addTweet: AddTweetDTO = new AddTweetDTO();


    if (this.isChecked == true) {

      addTweet.advertisement = true

      if (this.advertisementFormGroup.invalid) {
        return;
      }
    }else {
      addTweet.advertisement = false
    }
    addTweet.text = this.tweetFormGroup.get("text")?.value;

    console.log(addTweet)

    this.formData.append("json", JSON.stringify(addTweet))
      this.tweetService.AddTweet(this.formData).subscribe({
        next: (data: Tweet) => {

          if (data.advertisement){
            var adConfig: AdConfig = new AdConfig()
            adConfig.tweet_id = data.id
            adConfig.age_from = this.advertisementFormGroup.get("age_from")?.value
            adConfig.age_to = this.advertisementFormGroup.get("age_to")?.value   
            adConfig.gender = this.advertisementFormGroup.get("gender")?.value
            adConfig.residence = this.advertisementFormGroup.get("residence")?.value
            
            this.followService.CreateAdd(adConfig).subscribe()
          }


          this.router.navigate(['/Main-Page']);
        },
        error: (error) => {
          console.log(error);
        }
      })

      


    
  }

  getFile(event: any) {
    console.log("Desio se event")
    this.file = event.target.files[0];
    let fileType = this.file.type.split("/")
    if (fileType[0] === "image") {
        this.formData.append('image', this.file);
    } else {
        console.log('Wrong file type. Only jpeg images are allowed.');
    }
  }

  uploadFile() {
    this.http.post('/api/upload', this.formData).subscribe(response => {
      console.log(response);
    });
  }

}
