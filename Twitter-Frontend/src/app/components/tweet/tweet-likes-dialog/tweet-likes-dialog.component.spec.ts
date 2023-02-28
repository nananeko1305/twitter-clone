import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TweetLikesDialogComponent } from './tweet-likes-dialog.component';

describe('TweetLikesDialogComponent', () => {
  let component: TweetLikesDialogComponent;
  let fixture: ComponentFixture<TweetLikesDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TweetLikesDialogComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TweetLikesDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
