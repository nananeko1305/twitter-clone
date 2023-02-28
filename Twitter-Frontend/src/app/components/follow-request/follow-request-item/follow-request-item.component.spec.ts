import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FollowRequestItemComponent } from './follow-request-item.component';

describe('FollowRequestItemComponent', () => {
  let component: FollowRequestItemComponent;
  let fixture: ComponentFixture<FollowRequestItemComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FollowRequestItemComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FollowRequestItemComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
