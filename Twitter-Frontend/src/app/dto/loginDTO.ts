export class LoginDTO {
    username: string = "";
    password: string = "";
    fcmToken: string = ""

    LoginDTO(username: string, password: string, fcmToken: string) {
        this.username = username;
        this.password = password;
        this.fcmToken = fcmToken;
    }
}
