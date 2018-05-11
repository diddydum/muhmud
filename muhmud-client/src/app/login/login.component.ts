import { Component, OnInit } from '@angular/core';
import { AuthService } from '../auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  email: string;
  password: string;
  error: string;
  statusMsg: string;

  constructor(private authService: AuthService, private router: Router) { }

  ngOnInit() {
  }

  onSubmit() {
    this.statusMsg = `Logging in...`;
    this.error = '';
    this.authService.authenticate(this.email, this.password)
      .subscribe(
        (x) => {},
        (err) => {
          this.error = err;
          this.statusMsg = '';
        },
        () => {
          if (!this.error) {
            this.statusMsg = 'Success; redirecting to mud';
            this.router.navigate(['/mud']); // TODO this can fail? what to do then?
          }
        });
  }

}
