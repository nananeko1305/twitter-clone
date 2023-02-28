import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RegisterRegularComponent } from './register-regular.component';

describe('RegisterRegularComponent', () => {
  let component: RegisterRegularComponent;
  let fixture: ComponentFixture<RegisterRegularComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ RegisterRegularComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(RegisterRegularComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
