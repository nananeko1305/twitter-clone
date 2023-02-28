export class User {
    firstName: string = "";
    lastName: string = "";
    gender: string = "";
    age: number = 0;
    residence: string = "";
    username: string = "";
    password: string = "";
    companyName: string = "";
    email: string = "";
    website: string = "";
    userType: string = "";
    privacy: boolean = false;

    User(firstName: string, lastName: string, gender: string, age: number, residence: string, username: string, password: string, companyName: string, email: string, website: string, userType: string, privacy: boolean) {
        this.firstName = firstName;
        this.lastName = lastName;
        this.gender = gender;
        this.age = age;
        this.residence = residence;
        this.username = username;
        this.password = password;
        this.companyName = companyName;
        this.email = email;
        this.website = website;
        this.userType = userType;
        this.privacy = privacy
    }
}