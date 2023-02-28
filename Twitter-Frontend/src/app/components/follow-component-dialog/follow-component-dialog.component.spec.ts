import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FollowComponentDialogComponent } from './follow-component-dialog.component';

describe('FollowComponentDialogComponent', () => {
  let component: FollowComponentDialogComponent;
  let fixture: ComponentFixture<FollowComponentDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FollowComponentDialogComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FollowComponentDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
